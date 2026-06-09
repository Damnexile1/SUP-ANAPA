document.addEventListener('DOMContentLoaded', function() {
    loadInstructors();
    setupBookingFlow();
});

function loadInstructors() {
    fetch('/api/instructors')
        .then(function(res) { return res.json(); })
        .then(function(instructors) {
            var grid = document.getElementById('instructors-grid');
            if (!grid) return;

            grid.innerHTML = '';
            instructors = Array.isArray(instructors) ? instructors : [];

            if (instructors.length === 0) {
                grid.innerHTML = '<p class="text-gray-500 text-center col-span-full">Инструкторы скоро появятся</p>';
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
                div.className = 'bg-white rounded-2xl shadow-lg overflow-hidden hover:shadow-2xl transition duration-300';

                var img = document.createElement('img');
                img.src = photoURL;
                img.alt = instructor.Name || 'Инструктор SUP';
                img.className = 'w-full h-72 object-cover';
                img.onerror = function() { this.src = placeholder; };
                div.appendChild(img);

                var body = document.createElement('div');
                body.className = 'p-6';

                body.innerHTML =
                    '<h3 class="text-2xl font-bold text-gray-900 mb-2">' + (instructor.Name || 'Инструктор') + '</h3>' +
                    '<p class="text-gray-700 mb-3">' + (instructor.Description || 'Опытный инструктор SUP') + '</p>';

                if (instructor.Phone) {
                    body.innerHTML += '<p class="text-sm text-gray-500 mb-4">' + instructor.Phone + '</p>';
                }

                if (instructor.WalkTypes && instructor.WalkTypes.length > 0) {
                    var wtHtml = '<div class="text-sm text-gray-600 border-t border-gray-100 pt-4">' +
                        '<p class="font-semibold text-gray-900 mb-2">Прогулки:</p>';

                    instructor.WalkTypes.forEach(function(wt) {
                        wtHtml += '<p class="mb-1">• ' + wt.Name + ' — ' + wt.Price + ' ₽, до ' + wt.MaxPeople + ' чел.</p>';
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

            var grid = document.getElementById('instructors-grid');
            if (!grid) return;

            grid.innerHTML = '<p class="text-red-500 text-center col-span-full">Не удалось загрузить инструкторов</p>';
        });
}

var bookingState = {
    instructor: null,
    walkType: null,
    slot: null
};

function setupBookingFlow() {
    if (!document.getElementById('booking-instructors')) return;
    loadBookingInstructors();

    var form = document.getElementById('booking-form');
    if (!form) return;
    form.addEventListener('submit', submitBookingForm);
}

function loadBookingInstructors() {
    fetch('/api/instructors')
        .then(function(res) { return res.json(); })
        .then(function(instructors) {
            var container = document.getElementById('booking-instructors');
            if (instructors.length === 0) {
                container.innerHTML = '<p class="text-gray-500">Инструкторы пока недоступны</p>';
                return;
            }
            container.innerHTML = '';
            instructors.forEach(function(inst) {
                var card = document.createElement('button');
                card.type = 'button';
                card.className = 'text-left border rounded-lg p-4 hover:border-blue-500';
                card.innerHTML = '<p class="font-semibold">' + inst.Name + '</p><p class="text-sm text-gray-600">' + (inst.Description || '') + '</p>';
                card.onclick = function() { selectInstructor(inst); };
                container.appendChild(card);
            });
        });
}

function selectInstructor(inst) {
    bookingState.instructor = inst;
    bookingState.walkType = null;
    bookingState.slot = null;
    document.getElementById('walk-type-step').classList.remove('hidden');
    document.getElementById('slot-step').classList.add('hidden');
    document.getElementById('booking-form-container').classList.add('hidden');

    fetch('/api/instructors/' + inst.ID + '/walk-types')
        .then(function(res) { return res.json(); })
        .then(function(walkTypes) {
            var container = document.getElementById('walk-types-container');
            container.innerHTML = '';
            if (walkTypes.length === 0) {
                container.innerHTML = '<p class="text-gray-500">У инструктора пока нет типов прогулок</p>';
                return;
            }
            walkTypes.forEach(function(wt) {
                var card = document.createElement('button');
                card.type = 'button';
                card.className = 'text-left border rounded-lg p-4 hover:border-blue-500';
                card.innerHTML = '<p class="font-semibold">' + wt.Name + '</p>' +
                    '<p class="text-sm text-blue-700">' + wt.Price + ' ₽</p>' +
                    '<p class="text-sm text-gray-600">до ' + wt.MaxPeople + ' чел.</p>';
                card.onclick = function() { selectWalkType(wt); };
                container.appendChild(card);
            });
        });
}

function selectWalkType(walkType) {
    bookingState.walkType = walkType;
    bookingState.slot = null;
    document.getElementById('slot-step').classList.remove('hidden');
    document.getElementById('booking-form-container').classList.add('hidden');

    fetch('/api/slots?instructor_id=' + bookingState.instructor.ID + '&walk_type_id=' + walkType.ID)
        .then(function(res) { return res.json(); })
        .then(function(slots) {
            slots = Array.isArray(slots) ? slots : [];

            var container = document.getElementById('slots-container');
            container.innerHTML = '';

            if (slots.length === 0) {
                container.innerHTML = '<p class="text-gray-500">Нет доступных слотов для выбранной прогулки</p>';
                return;
            }

            var grouped = {};
            slots.forEach(function(slot) {
                if (slot.Status !== 'available') return;
                var date = new Date(slot.Date).toLocaleDateString('ru-RU');
                if (!grouped[date]) grouped[date] = [];
                grouped[date].push(slot);
            });

            Object.keys(grouped).sort().forEach(function(date) {
                var dateDiv = document.createElement('div');
                dateDiv.className = 'mb-4';
                dateDiv.innerHTML = '<h3 class="text-lg font-semibold mb-2">' + date + '</h3>';
                var grid = document.createElement('div');
                grid.className = 'grid grid-cols-1 md:grid-cols-2 gap-3';

                grouped[date].forEach(function(slot) {
                    var btn = document.createElement('button');
                    btn.type = 'button';
                    btn.className = 'border rounded-lg p-4 text-left hover:border-blue-500';
                    btn.innerHTML = '<p class="font-semibold">' + slot.StartTime.substring(0, 5) + ' - ' + slot.EndTime.substring(0, 5) + '</p>' +
                        '<p class="text-sm text-gray-600">' + slot.Price + ' ₽ • до ' + slot.MaxPeople + ' чел.</p>';
                    btn.onclick = function() {
                        console.log('Выбран слот:', slot);
                        selectSlot(slot, date);
                    };
                    grid.appendChild(btn);
                });

                dateDiv.appendChild(grid);
                container.appendChild(dateDiv);
            });
        });
}

function selectSlot(slot, dateLabel) {
    console.log('Выбран слот:', slot);

    bookingState.slot = slot;

    var selectedSlotInput = document.getElementById('selected-slot-id');
    var peopleCountInput = document.getElementById('people-count');
    var formContainer = document.getElementById('booking-form-container');
    var weatherInfo = document.getElementById('weather-info');

    if (!selectedSlotInput) {
        console.error('Не найден элемент #selected-slot-id');
        return;
    }

    if (!peopleCountInput) {
        console.error('Не найден элемент #people-count');
        return;
    }

    if (!formContainer) {
        console.error('Не найден элемент #booking-form-container');
        return;
    }

    if (!weatherInfo) {
        console.error('Не найден элемент #weather-info');
        return;
    }

    selectedSlotInput.value = slot.ID;
    peopleCountInput.max = slot.MaxPeople;

    var startTime = slot.StartTime ? slot.StartTime.substring(0, 5) : '';
    var endTime = slot.EndTime ? slot.EndTime.substring(0, 5) : '';

    weatherInfo.innerHTML =
        '<div class="bg-blue-50 border border-blue-200 rounded-lg p-4">' +
        '<h3 class="font-bold text-lg mb-3 text-blue-900">Вы выбрали прогулку</h3>' +
        '<div class="space-y-1 text-gray-800">' +
        '<p><strong>Инструктор:</strong> ' + bookingState.instructor.Name + '</p>' +
        '<p><strong>Прогулка:</strong> ' + (slot.WalkTypeName || bookingState.walkType.Name) + '</p>' +
        '<p><strong>Дата:</strong> ' + dateLabel + '</p>' +
        '<p><strong>Время:</strong> ' + startTime + ' - ' + endTime + '</p>' +
        '<p><strong>Цена:</strong> <span class="text-blue-700 font-bold">' + slot.Price + ' ₽</span></p>' +
        '<p><strong>Количество мест:</strong> до ' + slot.MaxPeople + ' чел.</p>' +
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
    var slotId = parseInt(formData.get('slot_id'));

    if (!bookingState.slot || !Number.isFinite(slotId) || slotId < 1) {
        document.getElementById('booking-result').innerHTML = '<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">' +
            '<p class="font-semibold">Выберите время прогулки</p>' +
            '<p class="text-sm">Сначала выберите инструктора, прогулку и доступный слот.</p></div>';
        return;
    }

    var data = {
        slot_id: slotId,
        client_name: formData.get('client_name'),
        client_phone: formData.get('client_phone'),
        client_email: formData.get('client_email'),
        people_count: parseInt(formData.get('people_count'))
    };

    var btn = form.querySelector('button[type="submit"]');
    btn.disabled = true;
    btn.textContent = 'Отправка...';

    fetch('/booking', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data)
    })
    .then(function(res) {
        if (!res.ok) {
            if (res.status === 401) {
                return res.text().then(function(err) {
                    throw new Error(err || 'Для бронирования заполните имя и телефон');
                });
            }
            return res.text().then(function(err) { throw new Error(err || 'Ошибка при бронировании'); });
        }
        return res.json();
    })
    .then(function(result) {
        document.getElementById('booking-result').innerHTML = '<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">' +
            '<p class="font-semibold">Отлично! Вы успешно забронировали прогулку 🎉</p>' +
            '<p class="text-sm"><strong>Номер бронирования:</strong> #' + result.ID + '</p>' +
            '<p class="text-sm"><strong>Маршрут:</strong> ' + bookingState.walkType.Name + '</p>' +
            // '<p class="text-sm">Статус и детали доступны в вашем <a href="/lk" class="underline font-semibold">личном кабинете</a>.</p>' +
            '<p class="text-sm">Администратор свяжется с вами для уточнения деталей</a>.</p>' +
            '<p class="text-sm mt-2"><strong>Статус:</strong> Ожидает подтверждения администратором в течение ' + result.hold_minutes + ' минут.</p>' +
            '</div>';
        form.reset();
        selectWalkType(bookingState.walkType);
    })
    .catch(function(err) {
        if (err && err.message === 'redirecting') return;
        document.getElementById('booking-result').innerHTML = '<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">' +
            '<p class="font-semibold">Ошибка при создании бронирования</p>' +
            '<p class="text-sm">' + (err.message || 'Пожалуйста, попробуйте еще раз') + '</p></div>';
    })
    .finally(function() {
        btn.disabled = false;
        btn.textContent = 'Забронировать';
    });
}
