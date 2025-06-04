// static/js/appointments.js

// Global variables
let apiBaseUrl;
let apptID;
let role;
let patientDetails = null;

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

// Load patient details from the API
async function loadPatientDetails() {
  try {
    console.log('Loading patient details for appointment:', apptID);
    const res = await fetch(`${apiBaseUrl}/api/appointments/${apptID}/details`);
    if (!res.ok) {
      console.error('Failed to load patient details:', res.status);
      return;
    }
    
    const data = await res.json();
    console.log('Patient details loaded:', data);
    patientDetails = data.patient;
    
    // Update the patient info in the UI
    updatePatientInfoUI();
  } catch (e) {
    console.error('loadPatientDetails error:', e);
  }
}

// Update the patient info section with data from the API
function updatePatientInfoUI() {
  if (!patientDetails) {
    console.error('No patient details available');
    return;
  }
  
  console.log('Updating UI with patient details:', patientDetails);
  
  // Update basic patient info
  document.querySelector('.info-grid').innerHTML = `
    <div class="label">ФИО</div>
    <div class="value">${patientDetails.full_name || 'Нет данных'}</div>
    <div class="label">ИИН</div>
    <div class="value">${patientDetails.iin || 'Нет данных'}</div>
    <div class="label">Telegram ID</div>
    <div class="value">${patientDetails.telegram_id || 'Нет данных'}</div>
  `;
  
  // Update medical history
  const historySection = document.querySelector('.info-section');
  
  // Medical conditions (chronic diseases)
  const medicalHistory = patientDetails.medical_history || [];
  const chronicHTML = medicalHistory.length > 0 
    ? `<ul>${medicalHistory.map(item => `<li>${item.diagnosis} ${item.date ? `(с ${item.date})` : ''}</li>`).join('')}</ul>` 
    : '<p>Нет данных</p>';
  
  // Allergies
  const allergies = patientDetails.allergies || [];
  const allergiesHTML = allergies.length > 0
    ? `<ul>${allergies.map(item => `<li>${item.name}</li>`).join('')}</ul>`
    : '<p>Нет данных</p>';
  
  // Vaccinations
  const vaccinations = patientDetails.vaccinations || [];
  const vaccinationsHTML = vaccinations.length > 0
    ? `<ul>${vaccinations.map(item => `<li>${item.vaccine} — ${item.date ? new Date(item.date).getFullYear() : 'дата неизвестна'}</li>`).join('')}</ul>`
    : '<p>Нет данных</p>';
  
  // Surgeries
  const surgeries = patientDetails.surgeries || [];
  const surgeriesHTML = surgeries.length > 0
    ? `<ul>${surgeries.map(item => `<li>${item.procedure} — ${item.date ? new Date(item.date).getFullYear() : 'дата неизвестна'}</li>`).join('')}</ul>`
    : '<p>Нет данных</p>';
  
  // Examinations
  const examinations = patientDetails.examinations || [];
  const examinationsHTML = examinations.length > 0
    ? `<ul>${examinations.map(item => `<li>${item.exam} — ${item.result} ${item.date ? `(${new Date(item.date).getFullYear()})` : ''}</li>`).join('')}</ul>`
    : '<p>Нет данных</p>';
  
  historySection.innerHTML = `
    <h3>Хронические заболевания</h3>
    ${chronicHTML}
    
    <h3>Аллергии</h3>
    ${allergiesHTML}
    
    <h3>Прививки</h3>
    ${vaccinationsHTML}
    
    <h3>Операции</h3>
    ${surgeriesHTML}
    
    <h3>Обследования</h3>
    ${examinationsHTML}
  `;
}

;(async function() {
  // Параметры из URL
  const params = new URLSearchParams(location.search);
  apptID = params.get('appointment_id') || params.get('id');
  role   = params.get('role');

  // в начале IIFE
  apiBaseUrl = window.location.origin;               // http(s)://domain.com

  try {
    const statusResponse = await fetch(`${apiBaseUrl}/api/appointments/${apptID}/status`);
    const statusData = await statusResponse.json();
    
    if (statusData.status === 'Завершен') {
      // Показать сообщение о завершенном приеме
      document.body.innerHTML = `
        <div style="display: flex; justify-content: center; align-items: center; height: 100vh; flex-direction: column; font-family: Arial, sans-serif;">
          <div style="text-align: center; padding: 40px; border-radius: 10px; box-shadow: 0 4px 6px rgba(0,0,0,0.1); background: white;">
            <h2 style="color: #666; margin-bottom: 20px;">📞 Прием завершен</h2>
            <p style="color: #888; margin-bottom: 30px;">Этот видеозвонок уже завершен и больше недоступен.</p>
            <button onclick="window.close()" style="background: #007bff; color: white; border: none; padding: 12px 24px; border-radius: 5px; cursor: pointer; font-size: 16px;">
              Закрыть
            </button>
          </div>
        </div>
      `;
      return; // Прерываем инициализацию
    }
  } catch (error) {
    console.error('Error checking appointment status:', error);
  }

  const wsProtocol = location.protocol === 'https:' 
                     ? 'wss:' 
                     : 'ws:';
  const wsUrl = `${wsProtocol}//${location.host}/ws?` +
                `appointment_id=${apptID}&role=${role}`;

  
  // Always load patient details regardless of role
  await loadPatientDetails();

  // Initialize chat panel
  const chatPanel = document.getElementById('chat-panel');
  const chatOverlay = document.getElementById('chatOverlay');
  const sidebar = document.querySelector('.sidebar');
  const toggleChatBtn = document.getElementById('toggleChat');
  const closeChatBtn = document.getElementById('closeChat');
  const videoArea = document.querySelector('.video-area')

  // Hide sidebar for patient, show only chat
  if (role === 'patient') {
  sidebar.style.display    = 'none';
  videoArea.style.display      = 'flex';
  toggleChatBtn.style.display  = 'inline-flex';
  chatPanel.classList.remove('open');
} else {
  sidebar.style.display    = 'flex';
  videoArea.style.display      = 'flex';
  toggleChatBtn.style.display  = 'inline-flex';
  chatPanel.classList.remove('open');
}

function openChat(){
  chatPanel.style.display = 'flex';
  chatPanel.classList.add('open');
  chatOverlay.classList.add('visible');

  if (role == 'patient') {
    videoArea.style.display = 'none';
  } else {
    sidebar.style.display = 'none';
  }
  loadChat();
}

function closeChat() {
  chatPanel.classList.remove('open');
  chatOverlay.classList.remove('visible');
  if (role == 'patient') {
    videoArea.style.display = 'flex';
  } else {
    sidebar.style.display = 'flex';
  }
}

  // Chat toggle button handler
  toggleChatBtn.onclick = () => {
    if (chatPanel.classList.contains('open')) closeChat();
    else openChat();
  };

  // Close chat button handler
  closeChatBtn.onclick = closeChat;
  chatOverlay.onclick = closeChat;

  // ===== WebRTC setup =====
  const pc = new RTCPeerConnection({
    iceServers: [
      // STUN серверы
      { urls: 'stun:stun.l.google.com:19302' },
      { urls: 'stun:stun.relay.metered.ca:80' },
      
      // TURN серверы с вашими credentials
      {
        urls: 'turn:a.relay.metered.ca:80',
        username: '47151ac2891d3c7c94c93235',
        credential: 'SzrCSFOALGQJO+da'
      },
      {
        urls: 'turn:a.relay.metered.ca:80?transport=tcp',
        username: '47151ac2891d3c7c94c93235',
        credential: 'SzrCSFOALGQJO+da'
      },
      {
        urls: 'turn:a.relay.metered.ca:443',
        username: '47151ac2891d3c7c94c93235',
        credential: 'SzrCSFOALGQJO+da'
      },
      {
        urls: 'turns:a.relay.metered.ca:443?transport=tcp',
        username: '47151ac2891d3c7c94c93235',
        credential: 'SzrCSFOALGQJO+da'
      }
    ]
  });
  

  pc.onicecandidate = e => {
    if (e.candidate) {
      console.log('🔗 ICE candidate:', e.candidate.type, e.candidate.candidate);
      if (e.candidate.type === 'relay') {
        console.log('✅ TURN сервер работает!');
      }
      if (ws.readyState === 1) {
        ws.send(JSON.stringify({ type: 'candidate', data: e.candidate }));
      }
    }
  };
  

  pc.ontrack = e => {
    document.getElementById('remote').srcObject = e.streams[0];
  };

  pc.oniceconnectionstatechange = () => {
    console.log('🌐 ICE connection state:', pc.iceConnectionState);
    
    if (pc.iceConnectionState === 'connected') {
      console.log('✅ Видеозвонок подключен успешно!');
    } else if (pc.iceConnectionState === 'failed') {
      console.log('❌ Не удалось подключиться. Проверьте TURN сервер.');
    } else if (pc.iceConnectionState === 'checking') {
      console.log('🔍 Проверка соединения...');
    }
  };
  
  pc.onconnectionstatechange = () => {
    console.log('🔌 Connection state:', pc.connectionState);
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
  ws.onerror = e => {
    console.error('❌ WebSocket error:', e); // Add this
  };
  ws.onclose = e => {
    console.log('🔌 WebSocket closed:', e.code, e.reason); // Add this
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

  document.getElementById('endCall').onclick = async () => {
    const confirmEnd = confirm('Вы уверены, что хотите завершить звонок?');
    if (!confirmEnd) return;
    
    try {
      console.log('🔴 Завершение звонка...');
      
      // 1. Закрываем WebRTC соединения
      pc.close();
      ws.close();
      
      // 2. Останавливаем локальный поток
      if (localStream) {
        localStream.getTracks().forEach(track => track.stop());
      }
      
      // 3. Отправляем запрос на завершение приема
      const response = await fetch(`${apiBaseUrl}/api/appointments/${apptID}/end-call`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' }
      });
      
      if (response.ok) {
        const data = await response.json();
        console.log('✅ Звонок завершен:', data);
        alert('✅ Звонок успешно завершен');
        closeAndGoBack();
      } else {
        const errorText = await response.text();
        console.error('❌ Server error:', errorText);
        alert('❌ Ошибка при завершении звонка: ' + errorText);
      }
    } catch (error) {
      console.error('❌ Error ending call:', error);
      alert('❌ Ошибка при завершении звонка');
    }
  };

  // Функция для закрытия страницы
  window.closeAndGoBack = function() {
    try {
      // Пытаемся закрыть текущую вкладку
      window.close();
      
      // Если браузер не позволяет закрыть, переходим назад
      setTimeout(() => {
        if (window.history.length > 1) {
          window.history.back();
        } else {
          // Если нет истории, переходим на главную
          window.location.href = '/';
        }
      }, 500);
    } catch (e) {
      // Fallback - переход на главную страницу
      window.location.href = '/';
    }
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
    const tabId = 'tab-' + btn.dataset.tab;
    document.getElementById(tabId).classList.add('active');
    
    // Load diagnoses when the diagnosis tab is clicked
    if (btn.dataset.tab === 'diagnosis') {
      loadDiagnoses();
    }
  });
});

// ===== Загрузка и обработка диагнозов =====
let diagnosesData = [];

// Load diagnoses when the page loads, not just when the tab is clicked
document.addEventListener('DOMContentLoaded', () => {
  // Load diagnoses with a slight delay to ensure all DOM elements are ready
  setTimeout(() => {
    loadDiagnoses();
  }, 500);
});

async function loadDiagnoses() {
  const diagnosesListEl = document.getElementById('diagnosesList');
  
  if (!diagnosesListEl) return; // Exit if element doesn't exist yet
  
  try {
    // Check if we already have diagnoses loaded
    if (diagnosesData.length > 0) {
      renderDiagnosesList(diagnosesData);
      return;
    }
    
    const response = await fetch(`${apiBaseUrl}/api/diagnoses`);
    
    if (!response.ok) {
      throw new Error(`HTTP error! Status: ${response.status}`);
    }
    
    const data = await response.json();
    
    if (data.success && Array.isArray(data.data)) {
      diagnosesData = data.data;
      renderDiagnosesList(diagnosesData);
    } else {
      diagnosesListEl.innerHTML = '<div class="loading">Ошибка при загрузке диагнозов</div>';
    }
  } catch (error) {
    console.error('Error loading diagnoses:', error);
    diagnosesListEl.innerHTML = '<div class="loading">Ошибка при загрузке диагнозов</div>';
  }
}

function renderDiagnosesList(diagnoses) {
  const diagnosesListEl = document.getElementById('diagnosesList');
  
  if (!diagnosesListEl) return; // Safety check
  
  if (!diagnoses || diagnoses.length === 0) {
    diagnosesListEl.innerHTML = '<div class="loading">Нет доступных диагнозов</div>';
    return;
  }
  
  let html = '';
  diagnoses.forEach(diagnosis => {
    html += `
      <div class="diagnosis-item" data-id="${diagnosis.id}" data-code="${diagnosis.code}" data-name="${diagnosis.name}">
        <span class="code">${diagnosis.code}</span>
        <span class="name">${diagnosis.name}</span>
      </div>
    `;
  });
  
  diagnosesListEl.innerHTML = html;
  
  // Add click handlers to diagnosis items
  document.querySelectorAll('.diagnosis-item').forEach(item => {
    item.addEventListener('click', () => {
      selectDiagnosis(item);
    });
  });
}

function selectDiagnosis(item) {
  // Get diagnosis data
  const id = item.getAttribute('data-id');
  const code = item.getAttribute('data-code');
  const name = item.getAttribute('data-name');
  
  // Store the diagnosis value
  const diagnosisInput = document.getElementById('diagnosis');
  if (diagnosisInput) {
    diagnosisInput.value = `${code} — ${name}`;
  }
  
  // Update the display
  const diagnosisDisplay = document.getElementById('diagnosisDisplay');
  if (diagnosisDisplay) {
    diagnosisDisplay.textContent = `${code} — ${name}`;
  }
  
  // Show the selected diagnosis section
  const selectedDiagnosis = document.getElementById('selectedDiagnosis');
  if (selectedDiagnosis) {
    selectedDiagnosis.style.display = 'block';
  }
  
  // Highlight the selected item
  document.querySelectorAll('.diagnosis-item').forEach(i => {
    i.classList.remove('selected');
  });
  item.classList.add('selected');
}

// Search functionality for diagnoses - updated for real-time filtering
const diagnosisSearchInput = document.getElementById('diagnosisSearch');
if (diagnosisSearchInput) {
  diagnosisSearchInput.addEventListener('input', function() {
    const searchTerm = this.value.toLowerCase();
    
    if (!diagnosesData.length) return;
    
    const filteredDiagnoses = diagnosesData.filter(diagnosis => 
      diagnosis.code.toLowerCase().includes(searchTerm) ||
      diagnosis.name.toLowerCase().includes(searchTerm)
    );
    
    renderDiagnosesList(filteredDiagnoses);
    
    // If search is empty, clear selection
    if (!searchTerm) {
      const selectedDiagnosis = document.getElementById('selectedDiagnosis');
      if (selectedDiagnosis) {
        selectedDiagnosis.style.display = 'none';
      }
    }
  });
}

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
// после всех функций и WebRTC-логаики
document.getElementById("saveBtn").onclick = async () => {
  const complaints = document.getElementById("complaint").value.trim();
  const diagnosis  = document.getElementById("diagnosis").value;   // код или имя
  const assignment = document.getElementById("assignText").value.trim();
  const prescriptions = Array.from(
    document.querySelectorAll("#prescriptionsTable tbody tr")
  ).map(tr => ({
    med:      tr.cells[0].textContent,
    dose:     tr.cells[1].textContent,
    schedule: tr.cells[2].textContent,
  }));

  const payload = {
    complaints,
    diagnosis,
    assignment,
    prescriptions,
  };

  const res = await fetch(
    `${apiBaseUrl}/api/appointments/${apptID}/details`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }
  );
  if (res.ok) {
    alert("Сохранено успешно");
  } else {
    alert("Ошибка: " + (await res.text()));
  }
};

