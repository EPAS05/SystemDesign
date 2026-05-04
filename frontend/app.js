const API_BASE = '/api';

let selectedNodeId = null;
let currentNodeForChildren = 2;

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

document.querySelectorAll('.tab-btn').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
        btn.classList.add('active');
        const tabId = btn.dataset.tab;
        document.getElementById(tabId + 'Tab').classList.add('active');
        if (tabId === 'navigator') refreshTreeAndRight();
        else if (tabId === 'products') showPlaceholder('products');
        else if (tabId === 'enums') showPlaceholder('enums');
        else if (tabId === 'units') loadUnits();
    });
});

function showPlaceholder(tabId) {
    const container = document.getElementById(tabId === 'products' ? 'productsList' : 'enumsList');
    container.innerHTML = '<div class="empty-message" style="text-align:center; padding:3rem;">пока нет</div>';
}

//  =================================================================================== главная
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
            for (let child of children) await renderTreeNode(child, childUl, level+1);
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
    document.getElementById('rightPanelTitle').innerHTML = `${nodeName} -> дочерние категории`;
    const children = await fetchJSON(`${API_BASE}/nodes/${nodeId}/children`);
    const container = document.getElementById('rightChildrenList');
    if (!children.length) {
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

document.getElementById('btnAddChild').onclick = async () => {
    if (!selectedNodeId) { alert('Выберите элемент'); return; }
    const newName = prompt('Введите название нового узла');
    if (!newName) return;
    await fetchJSON(`${API_BASE}/nodes`, { method: 'POST', body: JSON.stringify({ name: newName, parent_id: selectedNodeId }) });
    await refreshTreeAndRight();
};

document.getElementById('btnDeleteNode').onclick = async () => {
    if (!selectedNodeId || selectedNodeId === 2) { alert('Нельзя удалить корневой узел или ничего не выбрано'); return; }
    if (confirm('Удалить узел и всех потомков?')) {
        await fetchJSON(`${API_BASE}/nodes/${selectedNodeId}`, { method: 'DELETE' });
        selectedNodeId = null;
        document.getElementById('selectedInfo').innerHTML = 'Ничего не выбрано';
        await refreshTreeAndRight();
    }
};

document.getElementById('btnEditUnits').onclick = async () => {
    if (!selectedNodeId) { alert('Выберите элемент'); return; }
    try {
        const units = await fetchJSON(`${API_BASE}/units`);
        if (!units.length) { alert('Нет доступных единиц измерения'); return; }
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
            alert('Единица измерения обновлена');
            await refreshTreeAndRight();
        });
    } catch(e) { alert('Ошибка: ' + e.message); }
};

// =================================================================================== ЕИ
async function loadUnits() {
    const container = document.getElementById('unitsList');
    container.innerHTML = 'Загрузка...';
    try {
        let units = await fetchJSON(`${API_BASE}/units`);
        if (!Array.isArray(units)) units = [];
        if (!units.length) { container.innerHTML = '<div class="empty-message">Нет единиц</div>'; return; }
        let html = '<table><thead><tr><th>ID</th><th>Название</th><th>Множитель</th><th>Действия</th></tr></thead><tbody>';
        for (let u of units) {
            html += `<tr>
                <td>${u.ID}</td>
                <td>${u.Name}</td>
                <td>${u.Multiplier}</td>
                <td><button class="action-edit" data-id="${u.ID}">ред</button> <button class="action-delete" data-id="${u.ID}">удл</button></td>
            </tr>`;
        }
        html += '</tbody></table>';
        container.innerHTML = html;
        document.querySelectorAll('.action-edit').forEach(btn => btn.onclick = () => editUnit(parseInt(btn.dataset.id)));
        document.querySelectorAll('.action-delete').forEach(btn => btn.onclick = () => deleteUnit(parseInt(btn.dataset.id)));
    } catch(e) { container.innerHTML = `<div class="error">${e.message}</div>`; }
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
    if (confirm('Удалить единицу измерения?')) {
        await fetchJSON(`${API_BASE}/units/${id}`, { method: 'DELETE' });
        loadUnits();
    }
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

async function init() {
    await refreshTree();
    await loadChildrenRight(2, 'Изделия');
    loadUnits();
    showPlaceholder('products');
    showPlaceholder('enums');
}
init();