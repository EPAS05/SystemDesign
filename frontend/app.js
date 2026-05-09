const API_BASE = '/api';

let selectedNodeId = null;
let currentNodeForChildren = 2;

// ==================== МО
function showModal(title, formHtml, onSave) {
    const modal = document.getElementById('modal');
    const modalBody = document.getElementById('modalBody');
    modalBody.innerHTML = `<h3>${title}</h3>${formHtml}<div style="margin-top:1rem"><button id="modalSaveBtn" class="primary">Сохранить</button></div>`;
    modal.style.display = 'block';
    document.querySelector('.close').onclick = () => modal.style.display = 'none';
    window.onclick = (e) => { if (e.target === modal) modal.style.display = 'none'; };
    document.getElementById('modalSaveBtn').onclick = async () => {
        await onSave();
        modal.style.display = 'none';
    };
}

function showConfirmModal(title, message, onConfirm) {
    const modal = document.getElementById('modal');
    const modalBody = document.getElementById('modalBody');
    modalBody.innerHTML = `
        <h3>${title}</h3>
        <p>${message}</p>
        <div style="margin-top:1rem; display:flex; gap:0.5rem; justify-content:flex-end;">
            <button id="modalConfirmNo" class="secondary">Нет</button>
            <button id="modalConfirmYes" class="primary">Да</button>
        </div>
    `;
    modal.style.display = 'block';
    const closeModal = () => modal.style.display = 'none';
    document.querySelector('.close').onclick = closeModal;
    window.onclick = (e) => { if (e.target === modal) closeModal(); };
    document.getElementById('modalConfirmYes').onclick = () => {
        closeModal();
        if (onConfirm) onConfirm();
    };
    document.getElementById('modalConfirmNo').onclick = closeModal;
}

function showInfoModal(title, message) {
    const modal = document.getElementById('modal');
    const modalBody = document.getElementById('modalBody');
    modalBody.innerHTML = `
        <h3>${title}</h3>
        <p>${message}</p>
        <div style="margin-top:1rem; display:flex; justify-content:flex-end;">
            <button id="modalInfoOk" class="primary">OK</button>
        </div>
    `;
    modal.style.display = 'block';
    const closeModal = () => modal.style.display = 'none';
    document.querySelector('.close').onclick = closeModal;
    window.onclick = (e) => { if (e.target === modal) closeModal(); };
    document.getElementById('modalInfoOk').onclick = closeModal;
}

window.alert = function (msg) { showInfoModal('Сообщение', msg); };
window.confirm = function (msg) { showConfirmModal('Подтверждение', msg, () => { }); return true; };
window.prompt = function (msg, def) { showInfoModal('Ввод', msg); return def || ''; };

async function fetchJSON(url, options = {}) {
    const res = await fetch(url, {
        headers: { 'Content-Type': 'application/json' },
        ...options
    });
    if (!res.ok) {
        const errText = await res.text();
        throw new Error(`${res.status}: ${errText}`);
    }
    return res.json();
}

// ==================== Переключение 
document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
        btn.classList.add('active');
        const tabId = btn.dataset.tab;
        document.getElementById(tabId + 'Tab').classList.add('active');
        if (tabId === 'navigator') refreshTreeAndRight();
        else if (tabId === 'products') loadProducts();
        else if (tabId === 'enums') loadEnums();
        else if (tabId === 'units') loadUnits();
    });
});

// ================================================ НАВИГАТОР
async function renderTreeNode(node, parentUl, level) {
    const li = document.createElement('li');
    li.className = 'tree-node';
    const div = document.createElement('div');
    div.className = 'tree-node-content';
    if (selectedNodeId === node.ID) div.classList.add('selected');
    const expandSpan = document.createElement('span');
    expandSpan.className = 'node-expand';
    expandSpan.textContent = '>';
    expandSpan.onclick = async (e) => {
        e.stopPropagation();
        await loadChildrenRight(node.ID, node.Name);
        let childUl = li.querySelector('.children-placeholder');
        if (!childUl) {
            childUl = document.createElement('ul');
            childUl.className = 'children-placeholder';
            li.appendChild(childUl);
            const children = await fetchJSON(`${API_BASE}/nodes/${node.ID}/children`);
            if (Array.isArray(children)) {
                for (let child of children) await renderTreeNode(child, childUl, level + 1);
            }
            expandSpan.textContent = 'v';
        } else {
            if (childUl.style.display === 'none') {
                childUl.style.display = '';
                expandSpan.textContent = 'v';
            } else {
                childUl.style.display = 'none';
                expandSpan.textContent = '>';
            }
        }
    };
    const nameSpan = document.createElement('span');
    nameSpan.className = 'node-name';
    nameSpan.textContent = node.Name;
    nameSpan.onclick = (e) => {
        e.stopPropagation();
        selectNodeForEdit(node.ID);
        loadChildrenRight(node.ID, node.Name);
    };
    div.appendChild(expandSpan);
    div.appendChild(nameSpan);
    li.appendChild(div);
    parentUl.appendChild(li);
}

async function selectNodeForEdit(nodeId) {
    selectedNodeId = nodeId;
    const node = await fetchJSON(`${API_BASE}/nodes/${nodeId}`);
    document.getElementById('selectedInfo').innerHTML = `Выбран: ${node.Name} (ID=${nodeId})`;
    document.querySelectorAll('.tree-node-content').forEach(el => el.classList.remove('selected'));
    await refreshTree();
}

async function refreshTree() {
    const container = document.getElementById('treeContainer');
    container.innerHTML = '';
    const rootNode = await fetchJSON(`${API_BASE}/nodes/2`);
    const rootUl = document.createElement('ul'); rootUl.className = 'node-tree';
    await renderTreeNode(rootNode, rootUl, 0);
    container.appendChild(rootUl);
}

async function loadChildrenRight(nodeId, nodeName) {
    currentNodeForChildren = nodeId;
    document.getElementById('rightPanelTitle').innerHTML = `${nodeName} > дочерние категории`;
    const children = await fetchJSON(`${API_BASE}/nodes/${nodeId}/children`);
    const container = document.getElementById('rightChildrenList');
    if (!children || !children.length) {
        container.innerHTML = '<div class="empty-message">Нет дочерних элементов</div>';
        return;
    }
    let itemsHtml = '';
    for (let child of children) {
        itemsHtml += `
            <div class="child-item">
                <span class="child-name" data-id="${child.ID}">${child.Name}</span>
                <button class="select-child-btn" data-id="${child.ID}">Выбрать</button>
            </div>
        `;
    }
    container.innerHTML = `<div class="child-list">${itemsHtml}</div>`;
    document.querySelectorAll('.child-name').forEach(el => el.onclick = () => selectNodeForEdit(parseInt(el.dataset.id)));
    document.querySelectorAll('.select-child-btn').forEach(btn => btn.onclick = () => selectNodeForEdit(parseInt(btn.dataset.id)));
}

async function refreshTreeAndRight() {
    await refreshTree();
    if (currentNodeForChildren) {
        const node = await fetchJSON(`${API_BASE}/nodes/${currentNodeForChildren}`);
        await loadChildrenRight(currentNodeForChildren, node.Name);
    } else await loadChildrenRight(2, 'Изделия');
}

document.getElementById('btnAddChild').onclick = () => {
    if (!selectedNodeId) {
        showInfoModal('Ошибка', 'Выберите элемент');
        return;
    }
    const formHtml = `
        <div class="form-group">
            <label>Название нового узла</label>
            <input type="text" id="newNodeName" class="form-control" placeholder="Введите название" autofocus>
        </div>
    `;
    showModal('Создание узла', formHtml, async () => {
        const newName = document.getElementById('newNodeName').value.trim();
        if (!newName) {
            showInfoModal('Ошибка', 'Название не может быть пустым');
            return;
        }
        try {
            await fetchJSON(`${API_BASE}/nodes`, {
                method: 'POST',
                body: JSON.stringify({ name: newName, parent_id: selectedNodeId })
            });
            await refreshTreeAndRight();
            showInfoModal('Успех', 'Узел создан');
        } catch (e) {
            showInfoModal('Ошибка', e.message);
        }
    });
};

document.getElementById('btnDeleteNode').onclick = () => {
    if (!selectedNodeId || selectedNodeId === 2) {
        showInfoModal('Ошибка', 'Нельзя удалить корневой узел или ничего не выбрано');
        return;
    }
    showConfirmModal('Подтверждение удаления', 'Удалить узел и всех потомков?', async () => {
        try {
            await fetchJSON(`${API_BASE}/nodes/${selectedNodeId}`, { method: 'DELETE' });
            selectedNodeId = null;
            document.getElementById('selectedInfo').innerHTML = 'Ничего не выбрано';
            await refreshTreeAndRight();
            showInfoModal('Успех', 'Узел удалён');
        } catch (e) {
            showInfoModal('Ошибка', e.message);
        }
    });
};

document.getElementById('btnEditUnits').onclick = async () => {
    if (!selectedNodeId) {
        showInfoModal('Ошибка', 'Выберите элемент');
        return;
    }
    try {
        const units = await fetchJSON(`${API_BASE}/units`);
        if (!units || !units.length) {
            showInfoModal('Ошибка', 'Нет доступных единиц измерения');
            return;
        }
        const node = await fetchJSON(`${API_BASE}/nodes/${selectedNodeId}`);
        const currentUnitId = node.UnitID || '';
        let options = '<option value="">Без единицы</option>';
        units.forEach(u => {
            options += `<option value="${u.ID}" ${currentUnitId === u.ID ? 'selected' : ''}>${u.Name} (ID ${u.ID})</option>`;
        });
        const form = `<div class="form-group"><label>Выберите единицу измерения для узла</label><select id="unitSelect">${options}</select></div>`;
        showModal('Редактировать ЕИ узла', form, async () => {
            const newUnitId = document.getElementById('unitSelect').value;
            const body = newUnitId ? { unit_id: parseInt(newUnitId) } : { unit_id: null };
            await fetchJSON(`${API_BASE}/nodes/${selectedNodeId}/unit`, { method: 'PUT', body: JSON.stringify(body) });
            showInfoModal('Успех', 'Единица измерения обновлена');
            await refreshTreeAndRight();
        });
    } catch (e) {
        showInfoModal('Ошибка', e.message);
    }
};

// ================================================ ПРОДУКТЫ
async function getAvailableClasses() {
    const descendants = await fetchJSON(`${API_BASE}/nodes/2/descendants`);
    let nodes = Array.isArray(descendants) ? [...descendants] : [];
    nodes = nodes.filter(n => n.ID !== 2);
    return nodes;
}

async function loadProducts() {
    const container = document.getElementById('productsList');
    container.innerHTML = 'Загрузка...';
    try {
        const terminals = await fetchJSON(`${API_BASE}/nodes/2/terminal-descendants`);
        let allProducts = [];
        if (Array.isArray(terminals)) {
            for (let term of terminals) {
                const prods = await fetchJSON(`${API_BASE}/nodes/${term.ID}/products`);
                if (Array.isArray(prods)) allProducts.push(...prods);
            }
        }
        if (!allProducts.length) {
            container.innerHTML = '<div class="empty-message">Нет продуктов</div>';
            return;
        }
        let html = `<table><thead><tr><th>ID</th><th>Название</th><th>Класс</th><th>Тип ед.</th><th>Вес/м (т)</th><th>Длина (м)</th><th style="text-align:center">Действия</th></tr></thead><tbody>`;
        for (let p of allProducts) {
            html += `<tr>
                <td>${p.ID}</td>
                <td>${p.Name}</td>
                <td>${p.ClassNodeID}</td>
                <td>${p.UnitType || '-'}</td>
                <td>${p.WeightPerMeter !== undefined && p.WeightPerMeter !== null ? p.WeightPerMeter : '-'}</td>
                <td>${p.PieceLength !== undefined && p.PieceLength !== null ? p.PieceLength : '-'}</td>
                <td class="action-cell"><button class="btn-edit" data-id="${p.ID}">Ред</button><button class="btn-delete" data-id="${p.ID}">Удл</button></td>
            </tr>`;
        }
        html += '</tbody></table>';
        container.innerHTML = html;
        document.querySelectorAll('.btn-edit').forEach(btn => btn.onclick = () => editProduct(parseInt(btn.dataset.id)));
        document.querySelectorAll('.btn-delete').forEach(btn => btn.onclick = () => deleteProduct(parseInt(btn.dataset.id)));
    } catch (e) {
        container.innerHTML = `<div class="error">${e.message}</div>`;
    }
}

async function editProduct(id) {
    try {
        const product = await fetchJSON(`${API_BASE}/products/${id}`);
        let availableClasses = await getAvailableClasses();
        if (!Array.isArray(availableClasses)) availableClasses = [];
        const options = availableClasses.map(c => `<option value="${c.ID}" ${c.ID === product.ClassNodeID ? 'selected' : ''}>${c.Name} (ID=${c.ID})</option>`).join('');
        const form = `
            <div class="form-group"><label>Название</label><input id="prodName" value="${product.Name}"></div>
            <div class="form-group"><label>Класс</label><select id="prodClass">${options}</select></div>
            <div class="form-group"><label>Тип единицы (mass/length/piece)</label><input id="prodUnitType" value="${product.UnitType || ''}"></div>
            <div class="form-group"><label>Вес погонного метра (т/м)</label><input id="prodWeight" type="number" step="any" value="${product.WeightPerMeter !== undefined && product.WeightPerMeter !== null ? product.WeightPerMeter : ''}"></div>
            <div class="form-group"><label>Длина штуки (м)</label><input id="prodLength" type="number" step="any" value="${product.PieceLength !== undefined && product.PieceLength !== null ? product.PieceLength : ''}"></div>
        `;
        showModal('Редактировать продукт', form, async () => {
            const body = {
                name: document.getElementById('prodName').value,
                class_node_id: parseInt(document.getElementById('prodClass').value),
                unit_type: document.getElementById('prodUnitType').value || null,
                weight_per_meter: parseFloat(document.getElementById('prodWeight').value) || null,
                piece_length: parseFloat(document.getElementById('prodLength').value) || null
            };
            if (!body.name) {
                showInfoModal('Ошибка', 'Название обязательно');
                return;
            }
            await fetchJSON(`${API_BASE}/products/${id}`, { method: 'PUT', body: JSON.stringify(body) });
            loadProducts();
        });
    } catch (e) {
        showInfoModal('Ошибка', e.message);
    }
}

async function deleteProduct(id) {
    showConfirmModal('Подтверждение удаления', 'Удалить продукт?', async () => {
        await fetchJSON(`${API_BASE}/products/${id}`, { method: 'DELETE' });
        loadProducts();
    });
}

document.getElementById('btnCreateProduct').onclick = async () => {
    try {
        let availableClasses = await getAvailableClasses();
        if (!Array.isArray(availableClasses)) availableClasses = [];
        if (!availableClasses.length) {
            showInfoModal('Ошибка', 'Нет доступных классов. Сначала создайте узлы в навигаторе.');
            return;
        }
        const options = availableClasses.map(c => `<option value="${c.ID}">${c.Name} (ID=${c.ID})</option>`).join('');
        const form = `
            <div class="form-group"><label>Название продукта</label><input id="prodName" placeholder="Например, Балка 20Б1"></div>
            <div class="form-group"><label>Класс</label><select id="prodClass">${options}</select></div>
            <div class="form-group"><label>Тип единицы (mass/length/piece)</label><input id="prodUnitType" placeholder="mass, length или piece"></div>
            <div class="form-group"><label>Вес погонного метра (т/м)</label><input id="prodWeight" type="number" step="any"></div>
            <div class="form-group"><label>Длина штуки (м)</label><input id="prodLength" type="number" step="any"></div>
        `;
        showModal('Создать продукт', form, async () => {
            const body = {
                name: document.getElementById('prodName').value,
                class_node_id: parseInt(document.getElementById('prodClass').value),
                unit_type: document.getElementById('prodUnitType').value || null,
                weight_per_meter: parseFloat(document.getElementById('prodWeight').value) || null,
                piece_length: parseFloat(document.getElementById('prodLength').value) || null
            };
            if (!body.name) {
                showInfoModal('Ошибка', 'Название обязательно');
                return;
            }
            await fetchJSON(`${API_BASE}/products`, { method: 'POST', body: JSON.stringify(body) });
            loadProducts();
        });
    } catch (e) {
        showInfoModal('Ошибка', e.message);
    }
};

document.getElementById('btnRefreshProducts').onclick = () => loadProducts();

// ================================================ ЕДИНИЦЫ ИЗМЕРЕНИЯ
async function loadUnits() {
    const container = document.getElementById('unitsList');
    container.innerHTML = 'Загрузка...';
    try {
        let units = await fetchJSON(`${API_BASE}/units`);
        if (!Array.isArray(units)) units = [];
        if (!units.length) { container.innerHTML = '<div class="empty-message">Нет единиц</div>'; return; }
        let html = `<tr><thead><tr><th>ID</th><th>Название</th><th>Множитель</th><th style="text-align:center">Действия</th></tr></thead><tbody>`;
        for (let u of units) {
            html += `<table>
                <td>${u.ID}</td>
                <td>${u.Name}</td>
                <td>${u.Multiplier}</td>
                <td class="action-cell"><button class="btn-edit" data-id="${u.ID}">Ред</button><button class="btn-delete" data-id="${u.ID}">Удл</button></td>
            </tr>`;
        }
        html += '</tbody></td>';
        container.innerHTML = html;
        document.querySelectorAll('.btn-edit').forEach(btn => btn.onclick = () => editUnit(parseInt(btn.dataset.id)));
        document.querySelectorAll('.btn-delete').forEach(btn => btn.onclick = () => deleteUnit(parseInt(btn.dataset.id)));
    } catch (e) { container.innerHTML = `<div class="error">${e.message}</div>`; }
}

async function editUnit(id) {
    const unit = await fetchJSON(`${API_BASE}/units/${id}`);
    const form = `
        <div class="form-group"><label>Название</label><input id="unitName" value="${unit.Name}"></div>
        <div class="form-group"><label>Множитель</label><input id="unitMult" value="${unit.Multiplier}"></div>
    `;
    showModal('Редактировать единицу', form, async () => {
        const body = {
            name: document.getElementById('unitName').value,
            multiplier: parseFloat(document.getElementById('unitMult').value)
        };
        await fetchJSON(`${API_BASE}/units/${id}`, { method: 'PUT', body: JSON.stringify(body) });
        loadUnits();
    });
}

async function deleteUnit(id) {
    showConfirmModal('Подтверждение удаления', 'Удалить единицу измерения?', async () => {
        await fetchJSON(`${API_BASE}/units/${id}`, { method: 'DELETE' });
        loadUnits();
    });
}

document.getElementById('btnCreateUnit').onclick = () => {
    const form = `
        <div class="form-group"><label>Название</label><input id="unitName"></div>
        <div class="form-group"><label>Множитель</label><input id="unitMult" value="1"></div>
    `;
    showModal('Создать единицу', form, async () => {
        const body = {
            name: document.getElementById('unitName').value,
            multiplier: parseFloat(document.getElementById('unitMult').value)
        };
        await fetchJSON(`${API_BASE}/units`, { method: 'POST', body: JSON.stringify(body) });
        loadUnits();
    });
};

document.getElementById('btnRefreshUnits').onclick = () => loadUnits();

// ================================================ ПЕРЕЧИСЛЕНИЯ
async function loadEnums() {
    const container = document.getElementById('enumsList');
    container.innerHTML = 'Загрузка...';
    try {
        let enums = await fetchJSON(`${API_BASE}/enums`);
        if (!Array.isArray(enums)) enums = [];
        if (!enums.length) {
            container.innerHTML = '<div class="empty-message">Нет перечислений</div>';
            return;
        }
        let html = `<table><thead><tr><th>ID</th><th>Название</th><th>Тип</th><th>Значения</th><th style="text-align:center">Действия</th></tr></thead><tbody>`;
        for (let e of enums) {
            let values = [];
            try { values = await fetchJSON(`${API_BASE}/enums/${e.ID}/values`); } catch (e) { values = []; }
            const allValues = values.map(v => v.Value).join(', ');
            html += `<tr>
                <td>${e.ID}</td>
                <td>${e.Name}</td>
                <td>${e.TypeNodeID} (${e.TypeNodeID === 4 ? 'Числовые' : e.TypeNodeID === 5 ? 'Строковые' : 'Картинки'})</td>
                <td class="values-cell">${allValues || '-'}</td>
                <td class="action-cell">
                    <button class="btn-edit" data-id="${e.ID}">Ред</button>
                    <button class="btn-delete" data-id="${e.ID}">Удл</button>
                    <button class="btn-values" data-id="${e.ID}">Знач</button>
                </td>
            </tr>`;
        }
        html += '</tbody></table>';
        container.innerHTML = html;
        document.querySelectorAll('.btn-edit').forEach(btn => btn.onclick = () => editEnum(parseInt(btn.dataset.id)));
        document.querySelectorAll('.btn-delete').forEach(btn => btn.onclick = () => deleteEnum(parseInt(btn.dataset.id)));
        document.querySelectorAll('.btn-values').forEach(btn => btn.onclick = () => manageEnumValues(parseInt(btn.dataset.id)));
    } catch (e) {
        container.innerHTML = `<div class="error">${e.message}</div>`;
    }
}

async function editEnum(id) {
    const enumData = await fetchJSON(`${API_BASE}/enums/${id}`);
    const form = `
        <div class="form-group"><label>Название</label><input id="enumName" value="${enumData.Name}"></div>
        <div class="form-group"><label>Описание</label><textarea id="enumDesc" style="resize: none;">${enumData.Description || ''}</textarea></div>
        <div class="form-group"><label>Тип (4-числовые, 5-строковые, 6-картинки)</label><input id="enumType" value="${enumData.TypeNodeID}"></div>
    `;
    showModal('Редактировать перечисление', form, async () => {
        const body = {
            name: document.getElementById('enumName').value,
            description: document.getElementById('enumDesc').value || null,
            type_node_id: parseInt(document.getElementById('enumType').value)
        };
        await fetchJSON(`${API_BASE}/enums/${id}`, { method: 'PUT', body: JSON.stringify(body) });
        loadEnums();
    });
}

async function deleteEnum(id) {
    showConfirmModal('Подтверждение удаления', 'Удалить перечисление? Все его значения будут удалены.', async () => {
        await fetchJSON(`${API_BASE}/enums/${id}`, { method: 'DELETE' });
        loadEnums();
    });
}

async function manageEnumValues(enumId) {
    let values = [];
    try { values = await fetchJSON(`${API_BASE}/enums/${enumId}/values`); } catch (e) { values = []; }
    let listHtml = '<ul style="list-style:none; padding:0;">';
    for (let v of values) {
        listHtml += `<li style="display:flex; justify-content:space-between; margin-bottom:8px;">
                        <span><strong>${v.Value}</strong> (порядок ${v.SortOrder})</span>
                        <span>
                            <button class="edit-value-btn" data-id="${v.ID}" style="margin-right:8px;">Ред</button>
                            <button class="delete-value-btn" data-id="${v.ID}">Удл</button>
                        </span>
                     </li>`;
    }
    listHtml += '</ul>';
    const form = `
        <div style="margin-bottom:1rem;">
            <button id="addValueBtn" class="primary" style="margin-top:0.5rem;">Добавить значение</button>
        </div>
        <div id="valuesList">${listHtml}</div>
        <div style="margin-top:1rem;">
            <button id="reorderValuesBtn" class="secondary">Изменить порядок</button>
        </div>
    `;
    function refreshValuesModal() {
        manageEnumValues(enumId);
    }
    showModal(`Значения перечисления (ID ${enumId})`, form, async () => { });
    setTimeout(() => {
        const addBtn = document.getElementById('addValueBtn');
        if (addBtn) addBtn.onclick = () => addEnumValue(enumId, refreshValuesModal);
        document.querySelectorAll('.edit-value-btn').forEach(btn => btn.onclick = () => editEnumValue(parseInt(btn.dataset.id), refreshValuesModal));
        document.querySelectorAll('.delete-value-btn').forEach(btn => btn.onclick = () => deleteEnumValue(parseInt(btn.dataset.id), refreshValuesModal));
        const reorderBtn = document.getElementById('reorderValuesBtn');
        if (reorderBtn) reorderBtn.onclick = () => reorderEnumValues(enumId, refreshValuesModal);
    }, 50);
}

async function addEnumValue(enumId, callback) {
    const form = `<div class="form-group"><label>Значение</label><input id="enumValue" placeholder="Новое значение"></div>
                  <div class="form-group"><label>Порядок (оставьте пустым для авто)</label><input id="sortOrder" type="number" step="1"></div>`;
    showModal('Добавить значение', form, async () => {
        const value = document.getElementById('enumValue').value.trim();
        if (!value) {
            showInfoModal('Ошибка', 'Значение не может быть пустым');
            return;
        }
        const sortOrder = document.getElementById('sortOrder').value ? parseInt(document.getElementById('sortOrder').value) : null;
        const body = { value };
        if (sortOrder !== null) body.sort_order = sortOrder;
        await fetchJSON(`${API_BASE}/enums/${enumId}/values`, { method: 'POST', body: JSON.stringify(body) });
        if (callback) callback();
    });
}

async function editEnumValue(valueId, callback) {
    const val = await fetchJSON(`${API_BASE}/enums/values/${valueId}`);
    const form = `<div class="form-group"><label>Значение</label><input id="enumValue" value="${val.Value}"></div>
                  <div class="form-group"><label>Порядок</label><input id="sortOrder" value="${val.SortOrder}" type="number" step="1"></div>`;
    showModal('Редактировать значение', form, async () => {
        const newValue = document.getElementById('enumValue').value.trim();
        if (!newValue) {
            showInfoModal('Ошибка', 'Значение не может быть пустым');
            return;
        }
        const sortOrder = parseInt(document.getElementById('sortOrder').value);
        const body = { value: newValue, sort_order: isNaN(sortOrder) ? val.SortOrder : sortOrder };
        await fetchJSON(`${API_BASE}/enums/values/${valueId}`, { method: 'PUT', body: JSON.stringify(body) });
        if (callback) callback();
    });
}

async function deleteEnumValue(valueId, callback) {
    showConfirmModal('Подтверждение удаления', 'Удалить значение перечисления?', async () => {
        await fetchJSON(`${API_BASE}/enums/values/${valueId}`, { method: 'DELETE' });
        if (callback) callback();
    });
}

async function reorderEnumValues(enumId, callback) {
    let values = [];
    try { values = await fetchJSON(`${API_BASE}/enums/${enumId}/values`); } catch (e) { values = []; }
    const idsList = values.map(v => v.ID).join(', ');
    const form = `<div class="form-group"><label>Новый порядок ID значений (через запятую)</label>
                  <input id="valueIds" placeholder="Например: ${idsList}"></div>
                  <p class="hint">Текущие ID: ${idsList}</p>`;
    showModal('Переупорядочить значения', form, async () => {
        const input = document.getElementById('valueIds').value;
        const ids = input.split(',').map(s => parseInt(s.trim())).filter(n => !isNaN(n));
        if (ids.length !== values.length) {
            showInfoModal('Ошибка', `Введите ровно ${values.length} ID. Текущие ID: ${idsList}`);
            return;
        }
        await fetchJSON(`${API_BASE}/enums/${enumId}/values/reorder`, { method: 'POST', body: JSON.stringify({ value_ids: ids }) });
        if (callback) callback();
    });
}

document.getElementById('btnCreateEnum').onclick = () => {
    const form = `
        <div class="form-group"><label>Название</label><input id="enumName"></div>
        <div class="form-group"><label>Описание</label><textarea id="enumDesc" style="resize: none;"></textarea></div>
        <div class="form-group"><label>Тип (4-числовые, 5-строковые, 6-картинки)</label><input id="enumType" value="5"></div>
    `;
    showModal('Создать перечисление', form, async () => {
        const body = {
            name: document.getElementById('enumName').value,
            description: document.getElementById('enumDesc').value || null,
            type_node_id: parseInt(document.getElementById('enumType').value)
        };
        if (!body.name) {
            showInfoModal('Ошибка', 'Название обязательно');
            return;
        }
        await fetchJSON(`${API_BASE}/enums`, { method: 'POST', body: JSON.stringify(body) });
        loadEnums();
    });
};
document.getElementById('btnRefreshEnums').onclick = () => loadEnums();

async function init() {
    await refreshTree();
    await loadChildrenRight(2, 'Изделия');
    loadProducts();
    loadEnums();
    loadUnits();
}
init();