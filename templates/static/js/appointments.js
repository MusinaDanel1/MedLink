// static/js/appointments.js

// глобальные переменные
let apiBaseUrl;
let apptID;
let role;
let patientDetails = null;

// загрузка и отрисовка чата
async function loadChat() {
  const params = new URLSearchParams(window.location.search);
  const id = params.get("appointment_id") || params.get("id");
  console.log("LOAD CHAT, using appointment ID=", id);
  try {
    const res = await fetch(
      `${apiBaseUrl}/api/appointments/${id}/messages`
    );
    if (!res.ok) return;
    const messages = await res.json();
    const msgList = document.getElementById("msgList");
    msgList.innerHTML = messages
      .map((m) => {
        const isMe = m.sender.toLowerCase() === role.toLowerCase();
        let senderLabel = "Неизвестный";
        if (m.sender.toLowerCase() === "doctor") {
          senderLabel = isMe ? "Вы" : "Доктор";
        } else if (m.sender.toLowerCase() === "patient") {
          senderLabel = isMe ? "Вы" : "Пациент";
        } else if (m.sender.toLowerCase() === "bot") {
          senderLabel = "Система";
        }
        return `
        <div class="message ${isMe ? "me" : "other"}">
          <div class="bubble">
            <div class="sender">${senderLabel}</div>
            ${m.content}
          </div>
        </div>`;
      })
      .join("");
    msgList.scrollTop = msgList.scrollHeight;
  } catch (e) {
    console.error("loadChat error:", e);
  }
}

// загрузка и отрисовка данных пациента
async function loadPatientDetails() {
  const params = new URLSearchParams(window.location.search);
  const id = params.get("appointment_id") || params.get("id");
  console.log("LOAD PATIENT DETAILS, using appointment ID=", id);
  try {
    const res = await fetch(
      `${apiBaseUrl}/api/appointments/${id}/details`
    );
    if (!res.ok) {
      console.error("Failed to load patient details:", res.status);
      return;
    }
    const data = await res.json();
    console.log("Patient details loaded:", data);
    patientDetails = data.patient;
    updatePatientInfoUI();
  } catch (e) {
    console.error("loadPatientDetails error:", e);
  }
}

// отрисовка UI данных пациента
function updatePatientInfoUI() {
  if (!patientDetails) {
    console.error("No patient details available");
    return;
  }
  console.log("Updating UI with patient details:", patientDetails);
  document.querySelector(".info-grid").innerHTML = `
    <div class="label">ФИО</div>
    <div class="value">${patientDetails.full_name||"Нет данных"}</div>
    <div class="label">ИИН</div>
    <div class="value">${patientDetails.iin||"Нет данных"}</div>
    <div class="label">Telegram ID</div>
    <div class="value">${patientDetails.telegram_id||"Нет данных"}</div>
  `;
  const historySection = document.querySelector(".info-section");
  // … остальная отрисовка истории пациента …
}

// main IIFE — разбирает URL, запускает загрузки и WebRTC/чат
;(async function() {
  console.log("RAW location.search =", window.location.search);

  // 1) Читаем appointment_id и role в глобалы
  const params = new URLSearchParams(window.location.search);
  apptID = params.get("appointment_id") || params.get("id");
  role   = params.get("role");
  console.log("PARSED apptID =", apptID, "role =", role);
  if (!apptID || !role) {
    return alert("Не передан appointment_id или role");
  }

  // 2) root для API
  apiBaseUrl = window.location.origin;

  // 3) WS-URL для сигналинга
  const wsUrl =
    `${location.protocol.replace("http","ws")}//${location.host}` +
    `/ws?appointment_id=${apptID}&role=${role}`;

  // 4) Загрузить данные пациента
  await loadPatientDetails();

  // 5) Инициализировать чат/сайдбар
  const chatPanel  = document.getElementById("chat-panel");
  const sidebar    = document.querySelector(".sidebar");
  const toggleChat = document.getElementById("toggleChat");
  const closeChat  = document.getElementById("closeChat");

  if (role === "patient") {
    sidebar.style.display    = "none";
    chatPanel.style.display  = "flex";
    toggleChat.style.display = "none";
  } else {
    sidebar.style.display    = "flex";
    chatPanel.style.display  = "none";
    toggleChat.style.display = "inline-flex";
  }
  toggleChat.onclick = () => {
    if (chatPanel.style.display === "none") {
      chatPanel.style.display = "flex";
      sidebar.style.display   = "none";
      loadChat();
    } else {
      chatPanel.style.display = "none";
      sidebar.style.display   = "flex";
    }
  };
  closeChat.onclick = () => {
    chatPanel.style.display = "none";
    sidebar.style.display   = "flex";
  };

  // 6) Настройка WebRTC
  const pc = new RTCPeerConnection({
    iceServers: [{ urls: "stun:stun.l.google.com:19302" }]
  });

  // 7) Локальное медиа
  let localStream;
  try {
    localStream = await navigator.mediaDevices.getUserMedia({
      video: true, audio: true
    });
    document.getElementById("local").srcObject = localStream;
    localStream.getTracks().forEach((t) => pc.addTrack(t, localStream));
  } catch (e) {
    console.error("getUserMedia error:", e);
  }

  // 8) Сигналинг по WebSocket
  const ws = new WebSocket(wsUrl);
  ws.onopen = async () => {
    if (role === "doctor") {
      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      ws.send(JSON.stringify({ type: "offer", sdp: pc.localDescription }));
    }
  };
  ws.onmessage = async ({data}) => {
    const msg = JSON.parse(data);
    if (msg.type === "offer") {
      await pc.setRemoteDescription(msg.sdp);
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      ws.send(JSON.stringify({ type: "answer", sdp: pc.localDescription }));
    } else if (msg.type === "answer") {
      await pc.setRemoteDescription(msg.sdp);
    } else if (msg.type === "candidate") {
      try { await pc.addIceCandidate(msg.candidate); }
      catch(e){ console.warn("addIceCandidate error:", e); }
    }
  };
  ws.onerror = e => console.error("WS error:", e);
  ws.onclose = () => console.log("WS closed");

  // 9) ICE → WS
  pc.onicecandidate = ({candidate}) => {
    if (candidate && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({ type: "candidate", candidate }));
    }
  };

  // 10) Рисуем remote-поток
  pc.ontrack = (e) => {
    document.getElementById("remote").srcObject = e.streams[0];
  };

  // 11) Кнопки mic/cam/end
  document.getElementById("toggleMic").onclick = () => {
    if (!localStream) return;
    const t = localStream.getAudioTracks()[0];
    t.enabled = !t.enabled;
    document.getElementById("toggleMic").style.opacity = t.enabled?"1":"0.5";
  };
  document.getElementById("toggleCamera").onclick = () => {
    if (!localStream) return;
    const t = localStream.getVideoTracks()[0];
    t.enabled = !t.enabled;
    document.getElementById("toggleCamera").style.opacity = t.enabled?"1":"0.5";
  };
  document.getElementById("endCall").onclick = () => {
    pc.close();
    ws.close();
  };

  // 12) Чат: отправка сообщений
  async function sendMessage(txt) {
    try {
      await fetch(
        `${apiBaseUrl}/api/appointments/${apptID}/messages`,
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ sender: role, content: txt })
        }
      );
    } catch (e) {
      console.error("sendMessage error:", e);
    }
  }
  const msgInput = document.getElementById("msgInput");
  const msgBtn   = document.getElementById("sendBtn");
  msgBtn.onclick = async () => {
    const txt = msgInput.value.trim();
    if (!txt) return;
    await sendMessage(txt);
    msgInput.value = "";
    loadChat();
  };
  msgInput.addEventListener("keypress", e => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault(); msgBtn.click();
    }
  });

  // 13) Запускаем чат
  loadChat();
  setInterval(loadChat, 3000);

  // Скрыть крестик для пациента
  if (role === "patient") {
    closeChat.style.display = "none";
  }
})();

// ===== Переключение вкладок (таб-меню) =====
;(function() {
  const tabLinks = document.querySelectorAll(".tab-link");
  const tabPanes = document.querySelectorAll(".tab-pane");
  if (!tabLinks.length || !tabPanes.length) {
    console.warn("Tab toggler: no .tab-link or .tab-pane found");
    return;
  }
  tabLinks.forEach(link => {
    link.addEventListener("click", () => {
      // 1) Снять active везде
      tabLinks.forEach(l => l.classList.remove("active"));
      tabPanes.forEach(p => p.classList.remove("active"));
      // 2) Установить active на кнопку и нужную панель
      link.classList.add("active");
      const pane = document.getElementById("tab-" + link.dataset.tab);
      if (pane) pane.classList.add("active");

      // 3) Для вкладки «diagnosis» можно загрузить диагнозы
      if (link.dataset.tab === "diagnosis") {
        loadDiagnoses();
      }
    });
  });
})();

