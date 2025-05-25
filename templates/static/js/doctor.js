document.addEventListener("DOMContentLoaded", async () => {
  console.log("Doctor dashboard loaded");

  function parseLocalDateTime(isoString) {
    const dt = new Date(isoString);
  // Проверяем, нужна ли коррекция
  console.log('Original ISO string:', isoString);
  console.log('Parsed date:', dt);
  console.log('Local time:', dt.toLocaleString());
  return dt;
  }

  const calendarEl = document.getElementById("calendar");
  const doctorIdElement = document.getElementById("doctorId");

  let doctorId = doctorIdElement ? doctorIdElement.value : null;
  console.log("Doctor ID from element:", doctorId);

  if (!doctorId || isNaN(doctorId) || doctorId === "0") {
    console.error("Missing or invalid doctor ID");
    return;
  }

  // Переменная для хранения всех пациентов
  let allPatients = [];

  // 0. Load doctor details
  try {
    console.log(`Fetching doctor data from /api/doctors/${doctorId}`);
    const doctorRes = await fetch(`/api/doctors/${doctorId}`);
    if (!doctorRes.ok) {
      throw new Error(`Failed to fetch doctor data: ${doctorRes.status}`);
    }

    const doctor = await doctorRes.json();
    console.log("Doctor data received:", doctor);

    // Update UI with doctor name and specialization
    const nameElement = document.querySelector(".doctor-info h5");
    if (nameElement && nameElement.textContent.includes("Загрузка")) {
      nameElement.textContent = doctor.name;
    }

    const specElement = document.querySelector(".doctor-info p");
    if (specElement && specElement.textContent.includes("Загрузка")) {
      specElement.textContent = doctor.specialization;
    }

    const navNameElement = document.querySelector(".navbar-text");
    if (navNameElement && navNameElement.textContent.includes("Загрузка")) {
      navNameElement.textContent = doctor.name;
    }

    // Update doctor services list in sidebar
    const servicesList = document.getElementById("servicesList");
    if (doctor.services && doctor.services.length > 0) {
      console.log(
        `Found ${doctor.services.length} services for doctor:`,
        doctor.services
      );

      servicesList.innerHTML = "";

      doctor.services.forEach((service, index) => {
        const li = document.createElement("li");
        li.className = "list-group-item service-filter";
        li.textContent = service.name;
        li.dataset.id = service.id;

        if (index === 0) {
          li.classList.add("active");
        }

        li.addEventListener("click", () => {
          document
            .querySelectorAll("#servicesList .service-filter")
            .forEach((item) => item.classList.remove("active"));

          li.classList.add("active");
          calendar.refetchEvents();
        });

        servicesList.appendChild(li);
      });
    } else {
      console.warn("No services found for doctor");
      servicesList.innerHTML =
        '<li class="list-group-item text-muted">Нет доступных услуг</li>';
    }

    // Populate services in the schedule modal dropdown
    const svcSel = document.getElementById("serviceId");
    svcSel.innerHTML = "";

    if (doctor.services && doctor.services.length > 0) {
      doctor.services.forEach((service) => {
        svcSel.add(new Option(service.name, service.id));
      });
      console.log("Populated service dropdown with doctor services");
    } else {
      console.log("No doctor services found, fetching all services");
      const svcRes = await fetch("/api/services");
      const services = await svcRes.json();
      console.log("All services received:", services);
      services.forEach((s) => svcSel.add(new Option(s.name, s.id)));
    }
  } catch (error) {
    console.error("Error fetching doctor data:", error);

    try {
      const svcRes = await fetch("/api/services");
      const services = await svcRes.json();
      const svcSel = document.getElementById("serviceId");
      svcSel.innerHTML = "";
      services.forEach((s) => svcSel.add(new Option(s.name, s.id)));
    } catch (err) {
      console.error("Error fetching services:", err);
    }
  }

  // 2. загрузить пациентов
  async function loadPatients() {
    try {
      console.log("Loading patients from database...");
      const res = await fetch("/api/patients");
      
      if (!res.ok) {
        throw new Error(`Failed to fetch patients: ${res.status}`);
      }
      
      const patients = await res.json();
      console.log("Patients loaded:", patients);
      
      if (!Array.isArray(patients)) {
        console.error("Patients data is not an array:", patients);
        allPatients = [];
        updatePatientDropdown([]);
        return;
      }

      const validPatients = patients.filter(patient => {
        if (!patient) {
          console.warn("Found null/undefined patient:", patient);
          return false;
        }
        
        if (!patient.ID && !patient.id) {
          console.warn("Patient without ID:", patient);
          return false;
        }
        
        return true;
      }).map(patient => ({
        // Используем поля с заглавными буквами из вашего API
        id: patient.ID || patient.id,
        full_name: patient.FullName || patient.full_name || patient.name || 'Неизвестно',
        iin: patient.IIN || patient.iin || 'Не указан'
      }));
      
      console.log("Processed patients:", validPatients);

      // Сохраняем всех пациентов в глобальную переменную
      allPatients = validPatients;
      
      // Обновляем dropdown в модальном окне
      updatePatientDropdown(validPatients);
      
      // Обновляем поиск пациентов
      updatePatientSearch();
      
    } catch (error) {
      console.error("Error loading patients:", error);
      allPatients = [];
      updatePatientDropdown([]);
    }
  }

  // Функция для обновления dropdown пациентов
  function updatePatientDropdown(patients) {
    const sel = document.getElementById("patientSelect");
    sel.innerHTML = '<option value="">Выберите пациента...</option>';
    
    if (!Array.isArray(patients) || patients.length === 0) {
      const option = new Option("Нет доступных пациентов", "");
      option.disabled = true;
      sel.appendChild(option);
      return;
    }

    patients.forEach((patient) => {
    if (!patient || !patient.id) {
      console.warn("Skipping invalid patient:", patient);
      return;
    }
    
    const fullName = patient.full_name || 'Неизвестно';
    const iin = patient.iin || 'Не указан';
    
    const option = new Option(
      `${fullName} (ИИН: ${iin})`,
      patient.id
    );
    sel.appendChild(option);
  });
  }

  // Функция для обновления поиска пациентов
  // Функция для обновления поиска пациентов
function updatePatientSearch() {
  const searchInput = document.getElementById("patientSearch");
  
  // Убираем предыдущие обработчики, чтобы избежать дублирования
  const newSearchInput = searchInput.cloneNode(true);
  searchInput.parentNode.replaceChild(newSearchInput, searchInput);
  
  newSearchInput.addEventListener("input", (e) => {
    const searchTerm = e.target.value.toLowerCase();
    
    if (searchTerm.length === 0) {
      updatePatientDropdown(allPatients);
      return;
    }
    
    const filteredPatients = allPatients.filter((patient) => {
      if (!patient) return false;
      
      const fullName = patient.full_name || '';
      const iin = patient.iin || '';

      return (
        fullName.toLowerCase().includes(searchTerm) ||
        iin.includes(searchTerm)
      );
    });
    
    updatePatientDropdown(filteredPatients);
  });
}


  await loadPatients();

  // 3. инициализировать календарь
  const calendar = new FullCalendar.Calendar(calendarEl, {
    initialView: "timeGridWeek",
    timeZone: "local",
    locale: "ru",
    buttonText: {
      today: "Сегодня",
    },
    allDayText: "Весь день",
    slotLabelFormat: {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    },
    eventTimeFormat: {
      hour: '2-digit',
      minute: '2-digit',
      hour12: false
    },
    slotMinTime: '00:00:00',
    slotMaxTime: '24:00:00',
    events: fetchEvents,
    selectable: true,
    select: (info) => {
      // Проверяем, что есть активная услуга
      const activeService = document.querySelector(
        "#servicesList .service-filter.active"
      );
      
      if (!activeService) {
        alert("Пожалуйста, выберите услугу для записи");
        return;
      }

      // Находим соответствующий график для выбранного времени
      const selectedServiceId = +activeService.dataset.id;
      
      // Здесь нужно найти подходящий график (schedule) для этой услуги и времени
      findScheduleForTimeSlot(selectedServiceId, info.start, info.end)
        .then((schedule) => {
          if (!schedule) {
            alert("Нет доступного графика для выбранного времени и услуги");
            return;
          }
          
          // Заполняем форму
          document.getElementById("apptScheduleId").value = schedule.id;
          document.getElementById("apptStart").value = info.startStr.slice(0, 16);
          document.getElementById("apptEnd").value = info.endStr.slice(0, 16);
          
          // Очищаем поиск и выбор пациента
          document.getElementById("patientSearch").value = "";
          document.getElementById("patientSelect").value = "";
          updatePatientDropdown(allPatients);
          
          // Показываем модальное окно
          new bootstrap.Modal(document.getElementById("apptModal")).show();
        })
        .catch((error) => {
          console.error("Error finding schedule:", error);
          alert("Ошибка при поиске графика");
        });
    },
    eventClick: async (info) => {
      const ext = info.event.extendedProps;
      if (ext.canAccept) {
        if (confirm("Принять пациента?")) {
          const res = await fetch(
            `/api/appointments/${info.event.id}/accept`,
            {
              method: "PUT",
            }
          );
          if (res.ok) {
            const body = await res.json();
            info.event.setProp("backgroundColor", "gray");
            if (body.videoUrl) {
              info.event.setExtendedProp("videoUrl", body.videoUrl);
              window.open(body.videoUrl, "_blank");
            }
          } else {
            alert("Ошибка при приёме");
          }
        }
      } else if (ext.videoUrl) {
        window.open(ext.videoUrl, "_blank");
      }
    },
  });
  calendar.render();

  // Функция для поиска подходящего графика
  async function findScheduleForTimeSlot(serviceId, startTime, endTime) {
    try {
      const res = await fetch(`/api/schedules?doctorId=${doctorId}`);
      if (!res.ok) {
        throw new Error(`Failed to fetch schedules: ${res.status}`);
      }
      
      const schedules = await res.json();
      
      // Находим график для данной услуги, который покрывает выбранное время
      const suitableSchedule = schedules.find((schedule) => {
        if (schedule.service_id !== serviceId) return false;
        
        const scheduleStart = new Date(schedule.start_time);
        const scheduleEnd = new Date(schedule.end_time);
        
        return startTime >= scheduleStart && endTime <= scheduleEnd;
      });
      
      return suitableSchedule || null;
    } catch (error) {
      console.error("Error finding schedule:", error);
      return null;
    }
  }

  // 4. формы
  document.getElementById("scheduleForm").onsubmit = async (e) => {
    e.preventDefault();

    const serviceId = document.getElementById("serviceId").value;
    const serviceName =
      document.getElementById("serviceId").options[
        document.getElementById("serviceId").selectedIndex
      ].text;
    const startTime = document.getElementById("startTime").value;
    const endTime = document.getElementById("endTime").value;
    const color = document.getElementById("color").value;

    if (!serviceId || !startTime || !endTime) {
      alert("Пожалуйста, заполните все поля");
      return;
    }

    let actualServiceId = +serviceId;

    if (actualServiceId < 0) {
      try {
        const createServiceRes = await fetch("/api/services", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            doctorId: +doctorId,
            name: serviceName,
          }),
        });

        if (!createServiceRes.ok) {
          throw new Error(
            `Failed to create service: ${createServiceRes.status}`
          );
        }

        const newService = await createServiceRes.json();
        actualServiceId = newService.id || actualServiceId;
      } catch (error) {
        console.error("Error creating service:", error);
      }
    }

    try {
      const response = await fetch("/api/schedules", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          doctorId,
          serviceId: actualServiceId,
          start: startTime,
          end: endTime,
          color: color,
        }),
      });

      if (!response.ok) {
        throw new Error(`Failed to create schedule: ${response.status}`);
      }

      calendar.refetchEvents();

      bootstrap.Modal.getInstance(
        document.getElementById("scheduleModal")
      ).hide();

      document.getElementById("startTime").value = "";
      document.getElementById("endTime").value = "";

      location.reload();
    } catch (error) {
      console.error("Error creating schedule:", error);
      alert("Ошибка при создании графика. Пожалуйста, попробуйте еще раз.");
    }
  };

  document.getElementById("apptForm").onsubmit = async (e) => {
    e.preventDefault();

    const scheduleId = document.getElementById("apptScheduleId").value;
    const patientId = document.getElementById("patientSelect").value;
    const startTime = document.getElementById("apptStart").value;
    const endTime = document.getElementById("apptEnd").value;

    if (!patientId || !startTime || !endTime || !scheduleId) {
      alert("Пожалуйста, заполните все поля и выберите пациента");
      return;
    }

    console.log("Sending appointment data:");
    console.log("Start time input value:", startTime);
    console.log("End time input value:", endTime);
  
    const startISO = new Date(startTime).toISOString();
    const endISO = new Date(endTime).toISOString();
  
    console.log("Start ISO:", startISO);
    console.log("End ISO:", endISO);

    try {
      const response = await fetch("/api/appointments", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          scheduleId: +scheduleId,
          patientId: +patientId,
          start: startISO,
          end: endISO,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || `Failed to create appointment: ${response.status}`);
      }

      calendar.refetchEvents();

      bootstrap.Modal.getInstance(document.getElementById("apptModal")).hide();

      // Очищаем форму
      document.getElementById("apptStart").value = "";
      document.getElementById("apptEnd").value = "";
      document.getElementById("patientSearch").value = "";
      document.getElementById("patientSelect").value = "";
      
      alert("Пациент успешно записан!");
      
    } catch (error) {
      console.error("Error creating appointment:", error);
      alert(`Ошибка при создании записи: ${error.message}`);
    }
  };

  // 5. собрать события
  async function fetchEvents(fetchInfo, success, failure) {
    console.log("── fetchEvents called ──");
    console.log(" fetchInfo:", fetchInfo);
    console.log(" doctorId:", doctorId);

    if (!doctorId || isNaN(doctorId)) {
      console.warn(" invalid doctorId, returning empty");
      success([]);
      return;
    }

    try {
      const schRes = await fetch(`/api/schedules?doctorId=${doctorId}`);
      console.log(" schRes.ok =", schRes.ok, "status=", schRes.status);
      const schs = await schRes.json();
      console.log(" schedules from server:", schs);

      const activeServiceElement = document.querySelector(
        "#servicesList .service-filter.active"
      );
      const selectedServiceId = activeServiceElement
        ? +activeServiceElement.dataset.id
        : null;

      console.log(" selectedServiceId:", selectedServiceId);

      const filteredSchs = selectedServiceId
        ? schs.filter((s) => s.service_id === selectedServiceId)
        : [];

      console.log(" filtered schedules:", filteredSchs);

      if (filteredSchs.length === 0) {
        console.warn(" no schedules to show");
        success([]);
        return;
      }

      const ids = filteredSchs.map((s) => s.id).join("&scheduleIDs[]=");
      console.log(" → GET /api/appointments?scheduleIDs[]=" + ids);
      const apptRes = await fetch(`/api/appointments?scheduleIDs[]=${ids}`);
      console.log(" apptRes.ok =", apptRes.ok, "status=", apptRes.status);
      const appts = await apptRes.json();
      console.log(" appointments from server:", appts);

      const events = [];

      filteredSchs.forEach((s) => {
        const ev = {
          id: `sch-${s.id}`,
          title: `График #${s.id}`,
          start: parseLocalDateTime(s.start_time),
          end: parseLocalDateTime(s.end_time),
          display: "background",
          backgroundColor: s.color,
        };
        console.log(" push schedule-event:", ev);
        events.push(ev);
      });

      appts.forEach((a) => {
        const schedule = filteredSchs.find((s) => s.id === a.schedule_id);
        const evColor = a.status === "Принят" ? "gray" : "#e91e63";
        events.push({
          id: a.id,
          title: `${a.status} - ${a.patient_name || 'Пациент'}`,
          start: parseLocalDateTime(a.startTime),
          end: parseLocalDateTime(a.endTime),
          display: "auto",
          backgroundColor: evColor,
          borderColor: "#000",
          borderWidth: 1,
          textColor: "#fff",
          extendedProps: {
            canAccept: a.status === "Записан",
            videoUrl: a.video_url,
            patientName: a.patient_name,
          },
        });
      });

      console.log(" events before success():", events);
      success(events);
    } catch (error) {
      console.error("Error fetching events:", error);
      failure(error);
    }
  }
});
