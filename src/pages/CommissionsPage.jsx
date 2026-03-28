import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, Scale, Eye, Edit2 } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';
import Modal from '../components/UI/Modal';

export default function CommissionsPage() {
  const { commissions, updateCommission } = useData();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [selectedComm, setSelectedComm] = useState(null);
  const [editComm, setEditComm] = useState(null);
  const [editForm, setEditForm] = useState({ status: '', decision: '', description: '' });

  const filtered = commissions.filter(c => {
    const matchSearch = !search ||
      c.student_fio?.toLowerCase().includes(search.toLowerCase()) ||
      c.discipline_name?.toLowerCase().includes(search.toLowerCase());
    const matchStatus = !statusFilter || c.status === statusFilter;
    return matchSearch && matchStatus;
  });

  const statuses = ['Назначена', 'Проведена', 'Закрыта'];

  const openView = (c) => {
    setSelectedComm(c);
  };

  const openEdit = (c) => {
    setEditForm({ status: c.status, decision: c.decision, description: c.description });
    setEditComm(c);
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    await updateCommission(editComm.id, editForm);
    setEditComm(null);
  };

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">Комиссия</h2>
      </div>

      <div className="filters-bar">
        <div className="search-input">
          <Search size={16} />
          <input
            placeholder="Поиск по студенту или дисциплине..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">Все статусы</option>
          {statuses.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead>
            <tr>
              <th>Студент</th>
              <th>Группа</th>
              <th>Дисциплина</th>
              <th>Оценка</th>
              <th>Дата назначения</th>
              <th>Статус</th>
              <th>Решение</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((c, i) => (
                <motion.tr
                  key={c.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ delay: i * 0.03 }}
                >
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{c.student_fio}</td>
                  <td>{c.student_group}</td>
                  <td>{c.discipline_name}</td>
                  <td><StatusBadge status={c.grade} /></td>
                  <td>{c.assigned_date}</td>
                  <td><StatusBadge status={c.status} /></td>
                  <td style={{ maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {c.decision || '—'}
                  </td>
                  <td>
                    <div style={{ display: 'flex', gap: '4px', justifyContent: 'flex-end' }}>
                      <button className="btn-icon" onClick={() => openView(c)} title="Подробнее"><Eye size={16} /></button>
                      <button className="btn-icon" onClick={() => openEdit(c)} title="Редактировать"><Edit2 size={16} /></button>
                    </div>
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
        {filtered.length === 0 && (
          <div className="empty-state">
            <Scale size={48} />
            <p>Записи о комиссиях не найдены</p>
          </div>
        )}
      </motion.div>

      {/* View Modal */}
      <Modal isOpen={!!selectedComm} onClose={() => setSelectedComm(null)} title="Детали комиссии" wide>
        {selectedComm && (
          <div>
            <div className="detail-grid">
              <div className="detail-item">
                <span className="detail-label">Студент</span>
                <span className="detail-value">{selectedComm.student_fio}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Группа</span>
                <span className="detail-value">{selectedComm.student_group}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Дисциплина</span>
                <span className="detail-value">{selectedComm.discipline_name}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Оценка</span>
                <span className="detail-value"><StatusBadge status={selectedComm.grade} /></span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Дата назначения</span>
                <span className="detail-value">{selectedComm.assigned_date}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Статус</span>
                <span className="detail-value"><StatusBadge status={selectedComm.status} /></span>
              </div>
              <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                <span className="detail-label">Описание</span>
                <span className="detail-value">{selectedComm.description || '—'}</span>
              </div>
              <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                <span className="detail-label">Решение комиссии</span>
                <span className="detail-value">{selectedComm.decision || '—'}</span>
              </div>
            </div>
          </div>
        )}
      </Modal>

      {/* Edit Modal */}
      <Modal isOpen={!!editComm} onClose={() => setEditComm(null)} title="Обновление решения комиссии">
        {editComm && (
          <form onSubmit={handleUpdate}>
            <div style={{ marginBottom: '16px', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
              <strong>{editComm.student_fio}</strong> — {editComm.discipline_name}
            </div>
            <div className="form-group">
              <label>Статус комиссии</label>
              <select value={editForm.status} onChange={(e) => setEditForm(f => ({ ...f, status: e.target.value }))}>
                {statuses.map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Описание</label>
              <textarea
                rows={2}
                value={editForm.description}
                onChange={(e) => setEditForm(f => ({ ...f, description: e.target.value }))}
                placeholder="Описание заседания..."
              />
            </div>
            <div className="form-group">
              <label>Решение комиссии</label>
              <textarea
                rows={2}
                value={editForm.decision}
                onChange={(e) => setEditForm(f => ({ ...f, decision: e.target.value }))}
                placeholder="Решение..."
              />
            </div>
            <div className="modal-actions">
              <button type="button" className="btn btn-secondary" onClick={() => setEditComm(null)}>Отмена</button>
              <button type="submit" className="btn btn-primary">Сохранить</button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  );
}
