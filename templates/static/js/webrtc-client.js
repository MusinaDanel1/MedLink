// static/js/webrtc-client.js
(async () => {
    // 1) Извлекаем параметры
    const params = new URLSearchParams(location.search);
    const appointmentId = params.get("appointment_id");
    const role = params.get("role"); // "doctor" | "patient"
    if (!appointmentId || !role) {
      alert("Не передан appointment_id или role");
      return;
    }
  
    // 2) Селектим DOM
    const localVideo = document.getElementById("local");
    const remoteVideo = document.getElementById("remote");
    const toggleMicBtn = document.getElementById("toggleMic");
    const toggleCamBtn = document.getElementById("toggleCamera");
    const toggleChatBtn = document.getElementById("toggleChat");
    const endCallBtn = document.getElementById("endCall");
    const chatPanel = document.getElementById("chat-panel");
    const closeChatBtn = document.getElementById("closeChat");
    const msgList = document.getElementById("msgList");
    const msgInput = document.getElementById("msgInput");
    const sendBtn = document.getElementById("sendBtn");
  
    // 3) Подключаемся к сигналингу
    const protocol = location.protocol === "https:" ? "wss" : "ws";
    const signalingUrl =
      `${protocol}://${location.host}/ws` +
      `?appointment_id=${appointmentId}&role=${role}`;
    const signaling = new WebSocket(signalingUrl);
  
    // 4) Создаём RTCPeerConnection
    const pc = new RTCPeerConnection({
      iceServers: [
        { urls: "stun:stun.l.google.com:19302" },
        // добавить TURN по необходимости
      ]
    });
  
    // 5) (опционально) DataChannel для чата
    let dc;
    if (role === "doctor") {
      dc = pc.createDataChannel("chat");
      setupDataChannel();
    } else {
      pc.ondatachannel = (ev) => {
        dc = ev.channel;
        setupDataChannel();
      };
    }
  
    function setupDataChannel() {
      dc.onmessage = (ev) => {
        const msg = ev.data;
        const div = document.createElement("div");
        div.className = "message other";
        div.innerHTML = `<div class="bubble">${msg}</div>`;
        msgList.appendChild(div);
        msgList.scrollTop = msgList.scrollHeight;
      };
    }
  
    // 6) Обработчики ICE
    pc.onicecandidate = (ev) => {
      if (ev.candidate) {
        signaling.send(
          JSON.stringify({ type: "ice", candidate: ev.candidate })
        );
      }
    };
  
    // 7) Обработка удалённого потока
    pc.ontrack = (ev) => {
      // ev.streams[0] — это MediaStream
      remoteVideo.srcObject = ev.streams[0];
    };
  
    // 8) Захват local-media
    const localStream = await navigator.mediaDevices.getUserMedia({
      audio: true,
      video: true
    });
    localVideo.srcObject = localStream;
    localStream.getTracks().forEach((t) => pc.addTrack(t, localStream));
  
    // 9) Сигналинг — обмен SDP и ICE
    signaling.onmessage = async (ev) => {
      const msg = JSON.parse(ev.data);
      switch (msg.type) {
        case "offer":
          await pc.setRemoteDescription(msg);
          const answer = await pc.createAnswer();
          await pc.setLocalDescription(answer);
          signaling.send(JSON.stringify(pc.localDescription));
          break;
        case "answer":
          await pc.setRemoteDescription(msg);
          break;
        case "ice":
          try {
            await pc.addIceCandidate(msg.candidate);
          } catch (e) {
            console.warn("Error adding ICE candidate", e);
          }
          break;
      }
    };
  
    signaling.onopen = async () => {
      if (role === "doctor") {
        const offer = await pc.createOffer();
        await pc.setLocalDescription(offer);
        signaling.send(JSON.stringify(pc.localDescription));
      }
    };
  
    // 10) UI: микрофон/камера
    toggleMicBtn.onclick = () => {
      const track = localStream.getAudioTracks()[0];
      track.enabled = !track.enabled;
      toggleMicBtn.style.opacity = track.enabled ? "1" : "0.4";
    };
    toggleCamBtn.onclick = () => {
      const track = localStream.getVideoTracks()[0];
      track.enabled = !track.enabled;
      toggleCamBtn.style.opacity = track.enabled ? "1" : "0.4";
    };
  
    // 11) UI: чат
    toggleChatBtn.onclick = () => {
      chatPanel.style.display =
        chatPanel.style.display === "none" ? "flex" : "none";
    };
    closeChatBtn.onclick = () => (chatPanel.style.display = "none");
    sendBtn.onclick = () => {
      const text = msgInput.value.trim();
      if (!text || !dc || dc.readyState !== "open") return;
      // показываем своё
      const div = document.createElement("div");
      div.className = "message me";
      div.innerHTML = `<div class="bubble">${text}</div>`;
      msgList.appendChild(div);
      msgList.scrollTop = msgList.scrollHeight;
      // шлём
      dc.send(text);
      msgInput.value = "";
    };
  
    // 12) Завершение звонка
    endCallBtn.onclick = () => {
      pc.getSenders().forEach((s) => s.track && s.track.stop());
      pc.close();
      signaling.close();
      // можно сделать редирект или дать кнопку «Начать заново»
    };
  })();
  