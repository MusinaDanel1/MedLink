// static/js/appointments.js

// Global variables
let apiBaseUrl;
let apptID;
let role;

// Global chat functions
async function loadChat() {
  try {
    const res = await fetch(
      `${apiBaseUrl}/api/appointments/${apptID}/messages`
    );
    if (!res.ok) return;
    const messages = await res.json();
    const msgList = document.getElementById('msgList');
    msgList.innerHTML = messages.map(m => {
      const isMe = m.sender.toLowerCase() === role.toLowerCase();
      let senderLabel = 'Неизвестный';
      if (m.sender.toLowerCase() === 'doctor') {
        senderLabel = isMe ? 'Вы' : 'Доктор';
      } else if (m.sender.toLowerCase() === 'patient') {
        senderLabel = isMe ? 'Вы' : 'Пациент';
      } else if (m.sender.toLowerCase() === 'bot') {
        senderLabel = 'Система';
      }
      return `
        <div class="message ${isMe ? 'me' : 'other'}">
          <div class="bubble">
            <div class="sender">${senderLabel}</div>
            ${m.content}
          </div>
        </div>`;
    }).join('');
    msgList.scrollTop = msgList.scrollHeight;
  } catch (e) {
    console.error('loadChat error:', e);
  }
}

;(async function() {
  // Параметры из URL
  const params = new URLSearchParams(location.search);
  apptID = params.get('id') || '1';
  role = params.get('role') || 'patient';
  apiBaseUrl = `${location.protocol}//${location.hostname}:8080`;
  const wsUrl = `${location.protocol.replace('http','ws')}//` +
                `${location.hostname}:8080/ws?appointment_id=${apptID}&role=${role}`;

  // Initialize chat panel
  const chatPanel = document.getElementById('chat-panel');
  const sidebar = document.querySelector('.sidebar');
  const toggleChatBtn = document.getElementById('toggleChat');
  const closeChatBtn = document.getElementById('closeChat');

  // Hide sidebar for patient, show only chat
  if (role === 'patient') {
    if (sidebar) sidebar.style.display = 'none';
    if (chatPanel) chatPanel.style.display = 'flex';
    if (toggleChatBtn) toggleChatBtn.style.display = 'none';
  } else {
    // Doctor: show sidebar, hide chat by default
    if (sidebar) sidebar.style.display = 'flex';
    if (chatPanel) chatPanel.style.display = 'none';
    if (toggleChatBtn) toggleChatBtn.style.display = 'inline-flex';
  }

  // Chat toggle button handler
  toggleChatBtn.onclick = () => {
    console.log('Toggle chat clicked');
    if (chatPanel.style.display === 'none') {
      // Show chat, hide sidebar
      chatPanel.style.display = 'flex';
      sidebar.style.display = 'none';
      loadChat();
    } else {
      // Hide chat, show sidebar
      chatPanel.style.display = 'none';
      sidebar.style.display = 'block';
    }
  };

  // Close chat button handler
  closeChatBtn.onclick = () => {
    console.log('Close chat clicked');
    chatPanel.style.display = 'none';
    sidebar.style.display = 'block';
  };

  // ===== WebRTC setup =====
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
  });
  pc.onicecandidate = ({ candidate }) => {
    if (candidate && ws.readyState === 1) {
      ws.send(JSON.stringify({ type: 'candidate', data: candidate }));
    }
  };
  pc.ontrack = e => {
    document.getElementById('remote').srcObject = e.streams[0];
  };

  let localStream;
  try {
    localStream = await navigator.mediaDevices.getUserMedia({
      video: true,
      audio: true
    });
    document.querySelector('.local video').srcObject = localStream;
    localStream.getTracks().forEach(t => pc.addTrack(t, localStream));
  } catch (e) {
    console.error('getUserMedia error:', e);
  }

  // ===== WebSocket сигнализация =====
  const ws = new WebSocket(wsUrl);
  ws.onopen = async () => {
    if (role === 'doctor') {
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      ws.send(JSON.stringify({ type: 'offer', data: offer }));
    }
  };
  ws.onmessage = async ({ data }) => {
    const msg = JSON.parse(data);
    if (msg.type === 'offer') {
      await pc.setRemoteDescription(msg.data);
      const ans = await pc.createAnswer();
      await pc.setLocalDescription(ans);
      ws.send(JSON.stringify({ type: 'answer', data: ans }));
    } else if (msg.type === 'answer') {
      await pc.setRemoteDescription(msg.data);
    } else if (msg.type === 'candidate') {
      try {
        await pc.addIceCandidate(msg.data);
      } catch (e) {
        console.warn('addIceCandidate error:', e);
      }
    }
  };
  ws.onerror = e => console.error('WS error:', e);
  ws.onclose = () => console.log('WS closed');

  // ===== Кнопки управления микрофон/камера/завершить =====
  document.getElementById('toggleMic').onclick = () => {
    const track = localStream.getAudioTracks()[0];
    track.enabled = !track.enabled;
    document.getElementById('toggleMic').style.opacity =
      track.enabled ? 1 : 0.5;
  };
  document.getElementById('toggleCamera').onclick = () => {
    const track = localStream.getVideoTracks()[0];
    track.enabled = !track.enabled;
    document.getElementById('toggleCamera').style.opacity =
      track.enabled ? 1 : 0.5;
  };
  document.getElementById('endCall').onclick = () => {
    pc.close();
    ws.close();
  };

  // ===== Функции работы с чатом =====
  async function sendMessage(text) {
    try {
      const res = await fetch(
        `${apiBaseUrl}/api/appointments/${apptID}/messages`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ sender: role, content: text })
        }
      );
      return res.ok ? res.json() : null;
    } catch (e) {
      console.error('sendMessage error:', e);
      return null;
    }
  }

  // ===== Элементы чата =====
  const msgList = document.getElementById('msgList');
  const inp = document.getElementById('msgInput');
  const sendBtn = document.getElementById('sendBtn');

  // Отправка по клику
  sendBtn.addEventListener('click', async () => {
    const txt = inp.value.trim();
    if (!txt) return;
    await sendMessage(txt);
    inp.value = '';
    await loadChat();
  });

  // Отправка по Enter
  inp.addEventListener('keypress', e => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendBtn.click();
    }
  });

  // Первичная и периодическая загрузка чата
  loadChat();
  setInterval(loadChat, 3000);

  if (role === 'patient') {
    if (closeChatBtn) closeChatBtn.style.display = 'none';
  }
})();  // конец IIFE

// ===== Переключение вкладок (таб-меню) =====
document.querySelectorAll('.tab-link').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tab-link').forEach(b =>
      b.classList.remove('active')
    );
    document.querySelectorAll('.tab-pane').forEach(p =>
      p.classList.remove('active')
    );
    btn.classList.add('active');
    document
      .getElementById('tab-' + btn.dataset.tab)
      .classList.add('active');
  });
});

// ===== Логика «Рецептов» =====
const addBtn   = document.getElementById('addPrescriptionBtn');
const wrapper  = document.getElementById('prescriptionsWrapper');
const presTbody= document.querySelector('#prescriptionsTable tbody');
const medInput = document.getElementById('presMed');
const doseInput= document.getElementById('presDose');
const schedInput = document.getElementById('presSchedule');

addBtn.addEventListener('click', () => {
  const med   = medInput.value.trim();
  const dose  = doseInput.value.trim();
  const sched = schedInput.value.trim();
  if (!med || !dose || !sched) {
    alert('Пожалуйста, заполните все поля.');
    return;
  }
  const row = document.createElement('tr');
  row.innerHTML = `
    <td>${med}</td>
    <td>${dose}</td>
    <td>${sched}</td>
    <td><button class="remove-btn">❌</button></td>
  `;
  presTbody.appendChild(row);
  wrapper.style.display = 'block';
  medInput.value = doseInput.value = schedInput.value = '';
  medInput.focus();

  row.querySelector('.remove-btn').addEventListener('click', () => {
    row.remove();
    if (!presTbody.querySelector('tr')) {
      wrapper.style.display = 'none';
    }
  });
});

// ===== Сохранение приёма =====
document.getElementById('saveBtn').addEventListener('click', async () => {
  const data = {
    complaints: document.getElementById('complaint').value,
    diagnosis:  document.getElementById('diagnosis').value,
    assignment: document.getElementById('assignText').value,
    prescriptions: Array.from(
      document.querySelectorAll('#prescriptionsTable tbody tr')
    ).map(r => ({
      med:      r.children[0].textContent,
      dose:     r.children[1].textContent,
      schedule: r.children[2].textContent
    }))
  };
  await fetch(`/api/appointments/${apptID}/complete`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  });
  alert('Приём сохранён и отправлен пациенту');
});
