document.addEventListener('DOMContentLoaded', function() {
    loadInstructors();
    setupBookingFlow();
});

var bookingState = {
    instructor: null,
    walkType: null,
    slot: null
};

function escapeHtml(value) {
    return String(value || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

function formatPrice(value) {
    var number = Number(value || 0);
    return number.toLocaleString('ru-RU') + ' ₽';
}

function loadInstructors() {
    var grid = document.getElementById('instructors-grid');
    if (!grid) return;

    fetch('/api/instructors')
        .then(function(res) { return res.json(); })
        .then(function(instructors) {
            instructors = Array.isArray(instructors) ? instructors : [];
            grid.innerHTML = '';

            if (instructors.length === 0) {
                grid.innerHTML = '<p class="text-[#776958] text-center col-span-full">Инструкторы скоро появятся</p>';
                return;
            }

            var limit = parseInt(grid.getAttribute('data-limit'), 10);
            if (Number.isFinite(limit) && limit > 0) {
                instructors = instructors.slice(0, limit);
            }

            var placeholder = 'https://via.placeholder.com/400x400?text=Instructor';

            instructors.forEach(function(instructor) {
                var photoURL = instructor.Photo || placeholder;

                var div = document.createElement('div');
                div.className = 'beige-card rounded-2xl overflow-hidden hover:shadow-2xl transition duration-300';

                var img = document.createElement('img');
                img.src = photoURL;
                img.alt = instructor.Name || 'Инструктор SUP';
                img.className = 'w-full h-72 object-cover';
                img.onerror = function() { this.src = placeholder; };
                div.appendChild(img);

                var body = document.createElement('div');
                body.className = 'p-6';

                body.innerHTML =
                    '<h3 class="text-2xl font-bold text-[#3f3326] mb-2">' + escapeHtml(instructor.Name || 'Инструктор') + '</h3>' +
                    '<p class="text-[#776958] mb-3">' + escapeHtml(instructor.Description || 'Опытный инструктор SUP') + '</p>';

                if (instructor.Phone) {
                    body.innerHTML += '<p class="text-sm text-[#8a7a66] mb-4">' + escapeHtml(instructor.Phone) + '</p>';
                }

                if (instructor.WalkTypes && instructor.WalkTypes.length > 0) {
                    var wtHtml = '<div class="text-sm text-[#776958] border-t border-amber-950/10 pt-4">' +
                        '<p class="font-semibold text-[#3f3326] mb-2">Прогулки:</p>';

                    instructor.WalkTypes.forEach(function(wt) {
                        wtHtml += '<p class="mb-1">• ' + escapeHtml(wt.Name) + ' — ' + formatPrice(wt.Price) + ', до ' + escapeHtml(wt.MaxPeople) + ' чел.</p>';
                    });

                    wtHtml += '</div>';
                    body.innerHTML += wtHtml;
                }

                div.appendChild(body);
                grid.appendChild(div);
            });
        })
        .catch(function(err) {
            console.error('Ошибка загрузки инструкторов:', err);
            grid.innerHTML = '<p class="text-red-700 text-center col-span-full">Не удалось загрузить инструкторов</p>';
        });
}

function setupBookingFlow() {
    if (!document.getElementById('booking-instructors')) return;

    loadBookingInstructors();

    var form = document.getElementById('booking-form');
    if (!form) return;

    form.addEventListener('submit', submitBookingForm);
}

function loadBookingInstructors() {
    var container = document.getElementById('booking-instructors');
    if (!container) return;

    container.innerHTML = '<p class="text-[#776958]">Загружаем инструкторов...</p>';

    fetch('/api/instructors')
        .then(function(res) { return res.json(); })
        .then(function(instructors) {
            instructors = Array.isArray(instructors) ? instructors : [];
            container.innerHTML = '';

            if (instructors.length === 0) {
                container.innerHTML = '<p class="text-[#776958]">Инструкторы пока недоступны</p>';
                return;
            }

            instructors.forEach(function(inst) {
                var card = document.createElement('button');
                card.type = 'button';
                card.className = 'text-left beige-card rounded-2xl p-5 hover:border-[#8a5a2f] transition duration-200';

                card.innerHTML =
                    '<p class="text-lg font-bold text-[#3f3326] mb-2">' + escapeHtml(inst.Name || 'Инструктор') + '</p>' +
                    '<p class="text-sm text-[#776958]">' + escapeHtml(inst.Description || 'Опытный инструктор SUP') + '</p>';

                card.onclick = function() {
                    selectInstructor(inst);
                    markSelectedButton(container, card);
                };

                container.appendChild(card);
            });
        })
        .catch(function(err) {
            console.error('Ошибка загрузки инструкторов:', err);
            container.innerHTML = '<p class="text-red-700">Не удалось загрузить инструкторов</p>';
        });
}

function markSelectedButton(container, selectedCard) {
    var buttons = container.querySelectorAll('button');
    buttons.forEach(function(button) {
        button.classList.remove('ring-2', 'ring-[#8a5a2f]', 'bg-[#fff3df]');
    });

    selectedCard.classList.add('ring-2', 'ring-[#8a5a2f]', 'bg-[#fff3df]');
}

function selectInstructor(inst) {
    bookingState.instructor = inst;
    bookingState.walkType = null;
    bookingState.slot = null;

    var walkTypeStep = document.getElementById('walk-type-step');
    var slotStep = document.getElementById('slot-step');
    var formContainer = document.getElementById('booking-form-container');

    if (walkTypeStep) walkTypeStep.classList.remove('hidden');
    if (slotStep) slotStep.classList.add('hidden');
    if (formContainer) formContainer.classList.add('hidden');

    var container = document.getElementById('walk-types-container');
    if (!container) return;

    container.innerHTML = '<p class="text-[#776958]">Загружаем типы прогулок...</p>';

    fetch('/api/instructors/' + inst.ID + '/walk-types')
        .then(function(res) { return res.json(); })
        .then(function(walkTypes) {
            walkTypes = Array.isArray(walkTypes) ? walkTypes : [];
            container.innerHTML = '';

            if (walkTypes.length === 0) {
                container.innerHTML = '<p class="text-[#776958]">У инструктора пока нет типов прогулок</p>';
                return;
            }

            walkTypes.forEach(function(wt) {
                var card = document.createElement('button');
                card.type = 'button';
                card.className = 'text-left beige-card rounded-2xl p-5 hover:border-[#8a5a2f] transition duration-200';

                card.innerHTML =
                    '<p class="text-lg font-bold text-[#3f3326] mb-2">' + escapeHtml(wt.Name || 'Прогулка') + '</p>' +
                    '<p class="text-xl font-extrabold text-[#8a5a2f] mb-1">' + formatPrice(wt.Price) + '</p>' +
                    '<p class="text-sm text-[#776958]">до ' + escapeHtml(wt.MaxPeople) + ' чел.</p>';

                card.onclick = function() {
                    selectWalkType(wt);
                    markSelectedButton(container, card);
                };

                container.appendChild(card);
            });

            walkTypeStep.scrollIntoView({ behavior: 'smooth', block: 'start' });
        })
        .catch(function(err) {
            console.error('Ошибка загрузки типов прогулок:', err);
            container.innerHTML = '<p class="text-red-700">Не удалось загрузить типы прогулок</p>';
        });
}

function selectWalkType(walkType) {
    bookingState.walkType = walkType;
    bookingState.slot = null;

    var slotStep = document.getElementById('slot-step');
    var formContainer = document.getElementById('booking-form-container');

    if (slotStep) slotStep.classList.remove('hidden');
    if (formContainer) formContainer.classList.add('hidden');

    var container = document.getElementById('slots-container');
    if (!container) return;

    container.innerHTML = '<p class="text-[#776958]">Загружаем свободное время...</p>';

    fetch('/api/slots?instructor_id=' + bookingState.instructor.ID + '&walk_type_id=' + walkType.ID)
        .then(function(res) { return res.json(); })
        .then(function(slots) {
            slots = Array.isArray(slots) ? slots : [];
            container.innerHTML = '';

            var availableSlots = slots.filter(function(slot) {
                return slot.Status === 'available';
            });

            if (availableSlots.length === 0) {
                container.innerHTML = '<p class="text-[#776958]">Нет доступных слотов для выбранной прогулки</p>';
                return;
            }

            var grouped = {};
            availableSlots.forEach(function(slot) {
                var date = new Date(slot.Date).toLocaleDateString('ru-RU');
                if (!grouped[date]) grouped[date] = [];
                grouped[date].push(slot);
            });

            Object.keys(grouped).sort().forEach(function(date) {
                var dateDiv = document.createElement('div');
                dateDiv.className = 'mb-5';

                dateDiv.innerHTML =
                    '<h3 class="text-lg font-bold text-[#3f3326] mb-3">' + escapeHtml(date) + '</h3>';

                var grid = document.createElement('div');
                grid.className = 'grid grid-cols-1 md:grid-cols-2 gap-3';

                grouped[date].forEach(function(slot) {
                    var startTime = slot.StartTime ? slot.StartTime.substring(0, 5) : '';
                    var endTime = slot.EndTime ? slot.EndTime.substring(0, 5) : '';

                    var btn = document.createElement('button');
                    btn.type = 'button';
                    btn.className = 'text-left beige-card rounded-2xl p-5 hover:border-[#8a5a2f] transition duration-200';

                    btn.innerHTML =
                        '<p class="text-lg font-bold text-[#3f3326] mb-1">' + escapeHtml(startTime) + ' - ' + escapeHtml(endTime) + '</p>' +
                        '<p class="text-sm text-[#776958]">' + formatPrice(slot.Price) + ' • до ' + escapeHtml(slot.MaxPeople) + ' чел.</p>';

                    btn.onclick = function() {
                        selectSlot(slot, date);
                        markSelectedButton(grid, btn);
                    };

                    grid.appendChild(btn);
                });

                dateDiv.appendChild(grid);
                container.appendChild(dateDiv);
            });

            slotStep.scrollIntoView({ behavior: 'smooth', block: 'start' });
        })
        .catch(function(err) {
            console.error('Ошибка загрузки слотов:', err);
            container.innerHTML = '<p class="text-red-700">Не удалось загрузить доступные слоты</p>';
        });
}

function selectSlot(slot, dateLabel) {
    bookingState.slot = slot;

    var selectedSlotInput = document.getElementById('selected-slot-id');
    var peopleCountInput = document.getElementById('people-count');
    var formContainer = document.getElementById('booking-form-container');
    var weatherInfo = document.getElementById('weather-info');

    if (!selectedSlotInput || !peopleCountInput || !formContainer || !weatherInfo) {
        console.error('Не найдены элементы формы бронирования');
        return;
    }

    selectedSlotInput.value = slot.ID;
    peopleCountInput.max = slot.MaxPeople;

    var startTime = slot.StartTime ? slot.StartTime.substring(0, 5) : '';
    var endTime = slot.EndTime ? slot.EndTime.substring(0, 5) : '';

    weatherInfo.innerHTML =
        '<div>' +
        '<h3 class="font-bold text-xl mb-4 text-[#3f3326]">Вы выбрали прогулку</h3>' +
        '<div class="grid grid-cols-1 md:grid-cols-2 gap-3 text-[#3f3326]">' +
        '<p><strong>Инструктор:</strong> ' + escapeHtml(bookingState.instructor.Name) + '</p>' +
        '<p><strong>Прогулка:</strong> ' + escapeHtml(slot.WalkTypeName || bookingState.walkType.Name) + '</p>' +
        '<p><strong>Дата:</strong> ' + escapeHtml(dateLabel) + '</p>' +
        '<p><strong>Время:</strong> ' + escapeHtml(startTime) + ' - ' + escapeHtml(endTime) + '</p>' +
        '<p><strong>Цена:</strong> <span class="text-[#8a5a2f] font-extrabold">' + formatPrice(slot.Price) + '</span></p>' +
        '<p><strong>Количество мест:</strong> до ' + escapeHtml(slot.MaxPeople) + ' чел.</p>' +
        '</div>' +
        '</div>';

    formContainer.classList.remove('hidden');
    formContainer.style.display = 'block';

    setTimeout(function() {
        formContainer.scrollIntoView({ behavior: 'smooth', block: 'start' });
    }, 100);
}

function submitBookingForm(e) {
    e.preventDefault();

    var form = e.target;
    var formData = new FormData(form);
    var slotId = parseInt(formData.get('slot_id'), 10);
    var resultContainer = document.getElementById('booking-result');

    if (!bookingState.slot || !Number.isFinite(slotId) || slotId < 1) {
        resultContainer.innerHTML =
            '<div class="bg-red-50 border border-red-200 text-red-800 px-4 py-3 rounded-2xl">' +
            '<p class="font-semibold">Выберите время прогулки</p>' +
            '<p class="text-sm">Сначала выберите инструктора, прогулку и доступный слот.</p>' +
            '</div>';
        return;
    }

    var data = {
        slot_id: slotId,
        client_name: formData.get('client_name'),
        client_phone: formData.get('client_phone'),
        client_email: formData.get('client_email'),
        people_count: parseInt(formData.get('people_count'), 10)
    };

    var btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Отправка...';
    btn.classList.add('opacity-70', 'cursor-not-allowed');

    fetch('/booking', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
        .then(function(res) {
            if (!res.ok) {
                return res.text().then(function(err) {
                    throw new Error(err || 'Ошибка при бронировании');
                });
            }

            return res.json();
        })
        .then(function(result) {
            resultContainer.innerHTML =
                '<div class="bg-[#f0f7e8] border border-green-200 text-green-900 px-5 py-4 rounded-2xl">' +
                '<p class="font-bold text-lg">Отлично! Заявка на прогулку создана</p>' +
                '<p class="text-sm mt-1"><strong>Номер бронирования:</strong> #' + escapeHtml(result.ID) + '</p>' +
                '<p class="text-sm"><strong>Маршрут:</strong> ' + escapeHtml(bookingState.walkType.Name) + '</p>' +
                '<p class="text-sm mt-2">Администратор свяжется с вами для уточнения деталей.</p>' +
                '<p class="text-sm mt-2"><strong>Статус:</strong> ожидает подтверждения администратором в течение ' + escapeHtml(result.hold_minutes || 20) + ' минут.</p>' +
                '</div>';

            form.reset();

            if (bookingState.walkType) {
                selectWalkType(bookingState.walkType);
            }
        })
        .catch(function(err) {
            resultContainer.innerHTML =
                '<div class="bg-red-50 border border-red-200 text-red-800 px-5 py-4 rounded-2xl">' +
                '<p class="font-bold">Ошибка при создании бронирования</p>' +
                '<p class="text-sm mt-1">' + escapeHtml(err.message || 'Пожалуйста, попробуйте еще раз') + '</p>' +
                '</div>';
        })
        .finally(function() {
            btn.disabled = false;
            btn.textContent = 'Забронировать';
            btn.classList.remove('opacity-70', 'cursor-not-allowed');
        });
}