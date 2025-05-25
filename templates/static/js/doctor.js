document.addEventListener('DOMContentLoaded', async () => {
  console.log('Doctor dashboard loaded');
  
  function parseLocalDateTime(isoString) {
  const dt = new Date(isoString);
  // getTimezoneOffset возвращает разницу в минутах:
  // сколько минут нужно добавить к локальному, чтобы получить UTC.
  // Мы делаем наоборот — к UTC прибавляем offset, и результат
  // в локальной зоне окажется тем, что было в БД.
  dt.setTime(dt.getTime() + dt.getTimezoneOffset() * 60 * 1000);
  return dt;
  }
  const calendarEl = document.getElementById('calendar');
  const doctorIdElement = document.getElementById('doctorId');
  
  // Check if we have a valid doctorId
  let doctorId = doctorIdElement ? doctorIdElement.value : null;
  console.log('Doctor ID from element:', doctorId);
  
  // Ensure doctorId is a valid number
  if (!doctorId || isNaN(doctorId) || doctorId === "0") {
  console.error("Missing or invalid doctor ID");
  // You may want to handle this case differently
  return; // Prevent further execution 
  }
  
  // 0. Load doctor details
  try {
  console.log(`Fetching doctor data from /api/doctors/${doctorId}`);
  const doctorRes = await fetch(`/api/doctors/${doctorId}`);
  if (!doctorRes.ok) {
  throw new Error(`Failed to fetch doctor data: ${doctorRes.status}`);
  }
  
  const doctor = await doctorRes.json();
  console.log('Doctor data received:', doctor);
  
  // Update UI with doctor name and specialization if not already set by template
  const nameElement = document.querySelector('.doctor-info h5');
  if (nameElement && nameElement.textContent.includes('Загрузка')) {
  nameElement.textContent = doctor.name;
  }
  
  const specElement = document.querySelector('.doctor-info p');
  if (specElement && specElement.textContent.includes('Загрузка')) {
  specElement.textContent = doctor.specialization;
  }
  
  const navNameElement = document.querySelector('.navbar-text');
  if (navNameElement && navNameElement.textContent.includes('Загрузка')) {
  navNameElement.textContent = doctor.name;
  }
  
  // Update doctor services list in sidebar
  const servicesList = document.getElementById('servicesList');
  if (doctor.services && doctor.services.length > 0) {
  console.log(`Found ${doctor.services.length} services for doctor:`, doctor.services);
  doctor.services.forEach(service => {
  const li = document.createElement('li');
  li.className = 'list-group-item';
  li.textContent = service.name;
  li.dataset.id = service.id;
  li.addEventListener('click', () => {
  const isActive = li.classList.contains('active');
  if (isActive) {
  li.classList.remove('active');
  } else {
  li.classList.add('active');
  }
  calendar.refetchEvents();
  });
  servicesList.appendChild(li);
  });
  } else {
  console.warn('No services found for doctor');
  servicesList.innerHTML = '<li class="list-group-item text-muted">Нет доступных услуг</li>';
  }
  
  // Populate services in the schedule modal dropdown
  const svcSel = document.getElementById('serviceId');
  svcSel.innerHTML = ''; // Clear any existing options
  
  if (doctor.services && doctor.services.length > 0) {
  doctor.services.forEach(service => {
  svcSel.add(new Option(service.name, service.id));
  });
  console.log('Populated service dropdown with doctor services');
  } else {
  // Fallback to fetching all services if doctor services are not available
  console.log('No doctor services found, fetching all services');
  const svcRes = await fetch('/api/services');
  const services = await svcRes.json();
  console.log('All services received:', services);
  services.forEach(s => svcSel.add(new Option(s.name, s.id)));
  }
  
  } catch (error) {
  console.error('Error fetching doctor data:', error);
  
  // Fallback: fetch all services if doctor-specific ones fail
  try {
  const svcRes = await fetch('/api/services');
  const services = await svcRes.json();
  const svcSel = document.getElementById('serviceId');
  svcSel.innerHTML = ''; // Clear any existing options
  services.forEach(s => svcSel.add(new Option(s.name, s.id)));
  } catch (err) {
  console.error('Error fetching services:', err);
  }
  }
  
  // No longer need this separate services fetch since we get them with doctor data
  // or fall back to it if doctor data fetch fails
  
  // 2. загрузить пациентов
  async function loadPatients() {
  try {
  const res = await fetch('/api/patients');
  const pts = await res.json();
  const sel = document.getElementById('patientId');
  sel.innerHTML = '';
  pts.forEach(p => sel.add(new Option(p.full_name, p.id)));
  } catch (error) {
  console.error('Error loading patients:', error);
  }
  }
  await loadPatients();
  
  // 3. инициализировать календарь
  const calendar = new FullCalendar.Calendar(calendarEl, {
  initialView: 'timeGridWeek',
  timeZone: 'local',
  locale: 'ru',
        buttonText: {
          today: 'Сегодня'
        },
        allDayText: 'Весь день',
        slotLabelFormat: {
          hour: '2-digit',
          minute: '2-digit',
          hour12: false // Используем 24-часовой формат
        },
  events: fetchEvents,
  selectable: true,
  select: info => {
  document.getElementById('apptScheduleId').value = 0;
  document.getElementById('apptStart').value =
  info.startStr.slice(0, 16);
  document.getElementById('apptEnd').value =
  info.endStr.slice(0, 16);
  new bootstrap.Modal(
  document.getElementById('apptModal')
  ).show();
  },
  eventClick: async info => {
  const ext = info.event.extendedProps;
  if (ext.canAccept) {
  if (confirm('Принять пациента?')) {
  const res = await fetch(`/api/appointments/${info.event.id}/accept`, {
  method: 'PUT'
  });
  if (res.ok) {
  const body = await res.json();
  info.event.setProp('backgroundColor', 'gray');
  if (body.videoUrl) {
  info.event.setExtendedProp('videoUrl', body.videoUrl);
  window.open(body.videoUrl, '_blank');
  }
  } else {
  alert('Ошибка при приёме');
  }
  }
  } else if (ext.videoUrl) {
  window.open(ext.videoUrl, '_blank');
  }
  }
  ,
  events: fetchEvents
  });
  calendar.render();
  
  // 4. формы
  document.getElementById('scheduleForm').onsubmit = async e => {
  e.preventDefault();
  
  const serviceId = document.getElementById('serviceId').value;
  const serviceName = document.getElementById('serviceId').options[document.getElementById('serviceId').selectedIndex].text;
  const startTime = document.getElementById('startTime').value;
  const endTime = document.getElementById('endTime').value;
  const color = document.getElementById('color').value;
  
  if (!serviceId || !startTime || !endTime) {
  alert('Пожалуйста, заполните все поля');
  return;
  }
  
  // If we're dealing with a temporary service (ID < 0), we need to create it first
  let actualServiceId = +serviceId;
  
  if (actualServiceId < 0) {
  try {
  // Create the service first
  const createServiceRes = await fetch('/api/services', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
  doctorId: +doctorId,
  name: serviceName
  })
  });
  
  if (!createServiceRes.ok) {
  throw new Error(`Failed to create service: ${createServiceRes.status}`);
  }
  
  const newService = await createServiceRes.json();
  actualServiceId = newService.id || actualServiceId;
  
  } catch (error) {
  console.error('Error creating service:', error);
  // Continue with the negative ID and let the backend handle it
  }
  }
  
  try {
  const response = await fetch('/api/schedules', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
  doctorId,
  serviceId: actualServiceId,
  start: startTime,
  end: endTime,
  color: color
  })
  });
  
  if (!response.ok) {
  throw new Error(`Failed to create schedule: ${response.status}`);
  }
  
  calendar.refetchEvents();
  
  bootstrap.Modal.getInstance(
  document.getElementById('scheduleModal')
  ).hide();
  
  // Clear form inputs for next time
  document.getElementById('startTime').value = '';
  document.getElementById('endTime').value = '';
  } catch (error) {
  console.error('Error creating schedule:', error);
  alert('Ошибка при создании графика. Пожалуйста, попробуйте еще раз.');
  }
  };
  
  document.getElementById('apptForm').onsubmit = async e => {
  e.preventDefault();
  
  const scheduleId = document.getElementById('apptScheduleId').value;
  const patientId = document.getElementById('patientId').value;
  const startTime = document.getElementById('apptStart').value;
  const endTime = document.getElementById('apptEnd').value;
  
  if (!patientId || !startTime || !endTime) {
  alert('Пожалуйста, заполните все поля');
  return;
  }
  
  try {
  const response = await fetch('/api/appointments', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
  scheduleId: +scheduleId,
  patientId: +patientId,
  start: new Date(startTime).toISOString(),
  end: new Date(endTime).toISOString()
  })
  });
  
  if (!response.ok) {
  throw new Error(`Failed to create appointment: ${response.status}`);
  }
  
  calendar.refetchEvents();
  
  bootstrap.Modal.getInstance(
  document.getElementById('apptModal')
  ).hide();
  
  // Clear form inputs for next time
  document.getElementById('apptStart').value = '';
  document.getElementById('apptEnd').value = '';
  } catch (error) {
  console.error('Error creating appointment:', error);
  alert('Ошибка при создании записи. Пожалуйста, попробуйте еще раз.');
  }
  };
  
  document.getElementById('addPatient').onclick = async () => {
  const name = prompt('Имя пациента');
  if (!name) return;
  
  const iin = prompt('ИИН');
  if (!iin) return;
  
  try {
  const response = await fetch('/api/patients', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ full_name: name, iin, telegram_id: 0 })
  });
  
  if (!response.ok) {
  throw new Error(`Failed to create patient: ${response.status}`);
  }
  
  await loadPatients();
  } catch (error) {
  console.error('Error adding patient:', error);
  alert('Ошибка при добавлении пациента. Пожалуйста, попробуйте еще раз.');
  }
  };
  
  
  // 5. собрать события
  async function fetchEvents(fetchInfo, success, failure) {
  console.log('── fetchEvents called ──');
  console.log(' fetchInfo:', fetchInfo);
  console.log(' doctorId:', doctorId);
  
  if (!doctorId || isNaN(doctorId)) {
  console.warn(' invalid doctorId, returning empty');
  success([]);
  return;
  }
  
  // 1) Запрос расписаний
  console.log(` → GET /api/schedules?doctorId=${doctorId}`);
  const schRes = await fetch(`/api/schedules?doctorId=${doctorId}`);
  console.log(' schRes.ok =', schRes.ok, 'status=', schRes.status);
  const schs = await schRes.json();
  console.log(' schedules from server:', schs);
  
  // 2) Фильтр по услугам
  const selectedServices = Array.from(
  document.querySelectorAll('#servicesList .active')
  ).map(li => +li.dataset.id);
  const showAll = !selectedServices.length ||
  document.getElementById('btnToggleAll')
  .classList.contains('active');
  console.log(' selectedServices:', selectedServices, 'showAll=', showAll);
  
  const filteredSchs = showAll
  ? schs
  : schs.filter(s => selectedServices.includes(s.service_id));
  console.log(' filtered schedules:', filteredSchs);
  
  if (filteredSchs.length === 0) {
  console.warn(' no schedules to show');
  success([]);
  return;
  }
  
  // 3) Запрос встреч
  const ids = filteredSchs.map(s => s.id).join('&scheduleIDs[]=');
  console.log(' → GET /api/appointments?scheduleIDs[]=' + ids);
  const apptRes = await fetch(`/api/appointments?scheduleIDs[]=${ids}`);
  console.log(' apptRes.ok =', apptRes.ok, 'status=', apptRes.status);
  const appts = await apptRes.json();
  console.log(' appointments from server:', appts);
  
  // 4) Формируем события
  const events = [];
  
  filteredSchs.forEach(s => {
  const ev = {
  id:             `sch-${s.id}`,
  title:          `График #${s.id}`,
  start:          parseLocalDateTime(s.start_time),
  end:            parseLocalDateTime(s.end_time),
  display:        'background',
  backgroundColor: s.color
  };
  console.log(' push schedule-event:', ev);
  events.push(ev);
  });
  
  appts.forEach(a => {
  const schedule = filteredSchs.find(s => s.id === a.schedule_id);
  // Для «Записан» сделаем яркий розовый, для «Принят» — серый
  const evColor = a.status === 'Принят'
  ? 'gray'
  : '#e91e63'; // ярко-розовый
  events.push({
  id: a.id,
  title: a.status,
  start: parseLocalDateTime(a.startTime),
  end:  parseLocalDateTime(a.endTime),
  // явно block-событие с цветом и рамкой
  display:        'auto',
  backgroundColor: evColor,
  borderColor:     '#000',
  borderWidth:     1,
  textColor:       '#fff',
  extendedProps: {
  canAccept: a.status === 'Записан',
  videoUrl:  a.video_url
  }
  });
  });
  
  
  
  console.log(' events before success():', events);
  success(events);
  }
  
  });