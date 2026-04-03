import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, Scale, Eye, Edit2 } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';
import Modal from '../components/UI/Modal';

export default function CommissionsPage() {
  const { commissions, teachers, updateCommission } = useData();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [selectedComm, setSelectedComm] = useState(null);
  const [editComm, setEditComm] = useState(null);
  const [editForm, setEditForm] = useState({ status: '', decision: '', description: '', meeting_date: '', members: [], controlType: '', grade: '' });

  const filtered = commissions.filter(c => {
    const matchSearch = !search || c.student_fio?.toLowerCase().includes(search.toLowerCase()) ||
      c.discipline_name?.toLowerCase().includes(search.toLowerCase());
    const matchStatus = !statusFilter || c.status === statusFilter;
    return matchSearch && matchStatus;
  });

  const statuses = ['Назначена', 'Проведена', 'Закрыта'];

  const openEdit = (c) => {
    const memberIds = (c.members || []).map(m => m.teacher_id);
    setEditForm({
      status: c.status, decision: c.decision, description: c.description,
      meeting_date: c.meeting_date || '',
      members: memberIds, controlType: '', grade: ''
    });
    setEditComm(c);
  };

  const toggleEditMember = (tid) => {
    setEditForm(f => ({
      ...f,
      members: f.members.includes(tid) ? f.members.filter(id => id !== tid) : [...f.members, tid]
    }));
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    const members = editForm.members.map(tid => ({
      teacher_id: tid,
      role: 'Член комиссии'
    }));
    await updateCommission(editComm.id, {
      status: editForm.status, decision: editForm.decision,
      description: editForm.description, meeting_date: editForm.meeting_date, members
    });
    setEditComm(null);
  };

  const membersStr = (members) => {
    if (!members || members.length === 0) return '—';
    return members.map(m => {
      const short = m.fio.split(' ').map((w, i) => i === 0 ? w : w[0] + '.').join(' ');
      return short;
    }).join(', ');
  };

  return (
    <div className="page">
      <div className="page-header"><h2 className="page-title">Комиссия</h2></div>
      <div className="filters-bar">
        <div className="search-input"><Search size={16} />
          <input placeholder="Поиск по студенту или дисциплине..." value={search} onChange={(e) => setSearch(e.target.value)} /></div>
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">Все статусы</option>
          {statuses.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead><tr>
            <th>Студент</th><th>Группа</th><th>Дисциплина</th><th>Дата назначения</th>
            <th>Дата проведения</th><th>Статус</th><th>Члены комиссии</th><th>Решение</th><th></th>
          </tr></thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((c, i) => (
                <motion.tr key={c.id} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} transition={{ delay: i * 0.03 }}>
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{c.student_fio}</td>
                  <td>{c.student_group}</td>
                  <td>{c.discipline_name}</td>
                  <td>{c.assigned_date}</td>
                  <td>{c.meeting_date || '—'}</td>
                  <td><StatusBadge status={c.status} /></td>
                  <td style={{ maxWidth: '200px', fontSize: '0.8rem', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {membersStr(c.members)}
                  </td>
                  <td style={{ maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{c.decision || '—'}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '4px', justifyContent: 'flex-end' }}>
                      <button className="btn-icon" onClick={() => setSelectedComm(c)} title="Подробнее"><Eye size={16} /></button>
                      <button className="btn-icon" onClick={() => openEdit(c)} title="Редактировать"><Edit2 size={16} /></button>
                    </div>
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
        {filtered.length === 0 && (<div className="empty-state"><Scale size={48} /><p>Записи о комиссиях не найдены</p></div>)}
      </motion.div>

      {/* View Modal */}
      <Modal isOpen={!!selectedComm} onClose={() => setSelectedComm(null)} title="Детали комиссии" wide>
        {selectedComm && (
          <div>
            <div className="detail-grid">
              <div className="detail-item"><span className="detail-label">Студент</span><span className="detail-value">{selectedComm.student_fio}</span></div>
              <div className="detail-item"><span className="detail-label">Группа</span><span className="detail-value">{selectedComm.student_group}</span></div>
              <div className="detail-item"><span className="detail-label">Дисциплина</span><span className="detail-value">{selectedComm.discipline_name}</span></div>
              <div className="detail-item"><span className="detail-label">Оценка</span><span className="detail-value"><StatusBadge status={selectedComm.grade} /></span></div>
              <div className="detail-item"><span className="detail-label">Дата назначения</span><span className="detail-value">{selectedComm.assigned_date}</span></div>
              <div className="detail-item"><span className="detail-label">Дата проведения</span><span className="detail-value">{selectedComm.meeting_date || '—'}</span></div>
              <div className="detail-item"><span className="detail-label">Статус</span><span className="detail-value"><StatusBadge status={selectedComm.status} /></span></div>
              <div className="detail-item" style={{ gridColumn: '1 / -1' }}><span className="detail-label">Описание</span><span className="detail-value">{selectedComm.description || '—'}</span></div>
              <div className="detail-item" style={{ gridColumn: '1 / -1' }}><span className="detail-label">Решение комиссии</span><span className="detail-value">{selectedComm.decision || '—'}</span></div>
              <div className="detail-item" style={{ gridColumn: '1 / -1' }}>
                <span className="detail-label">Состав комиссии</span>
                <div className="detail-value">
                  {selectedComm.members?.length > 0 ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '4px' }}>
                      {selectedComm.members.map((m, i) => (
                        <span key={i} style={{ fontSize: '0.875rem' }}>
                          {m.fio}
                        </span>
                      ))}
                    </div>
                  ) : '—'}
                </div>
              </div>
            </div>
          </div>
        )}
      </Modal>

      {/* Edit Modal */}
      <Modal isOpen={!!editComm} onClose={() => setEditComm(null)} title="Редактирование комиссии" wide>
        {editComm && (
          <form onSubmit={handleUpdate}>
            <div style={{ marginBottom: '16px', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
              <strong>{editComm.student_fio}</strong> — {editComm.discipline_name}
            </div>
            <div className="form-row">
              <div className="form-group"><label>Статус комиссии</label>
                <select value={editForm.status} onChange={(e) => setEditForm(f => ({ ...f, status: e.target.value }))}>
                  {statuses.map(s => <option key={s} value={s}>{s}</option>)}
                </select></div>
              <div className="form-group"><label>Дата проведения</label>
                <input type="date" value={editForm.meeting_date} onChange={(e) => setEditForm(f => ({ ...f, meeting_date: e.target.value }))} /></div>
            </div>
            <div className="form-group"><label>Состав комиссии</label>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', padding: '8px', background: 'var(--bg-tertiary)', borderRadius: '8px', maxHeight: '160px', overflowY: 'auto' }}>
                {teachers.map(t => (
                  <label key={t.id} style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '0.875rem' }}>
                    <input type="checkbox" checked={editForm.members.includes(t.id)} onChange={() => toggleEditMember(t.id)} />
                    {t.fio}
                  </label>
                ))}
              </div></div>
            <div className="form-group"><label>Описание</label>
              <textarea rows={2} value={editForm.description} onChange={(e) => setEditForm(f => ({ ...f, description: e.target.value }))} placeholder="Описание заседания..." /></div>
            <div className="form-group"><label>Решение комиссии</label>
              <textarea rows={2} value={editForm.decision} onChange={(e) => setEditForm(f => ({ ...f, decision: e.target.value }))} placeholder="Решение..." /></div>
            <div className="form-row">
              <div className="form-group"><label>Форма контроля</label>
                <select value={editForm.controlType} onChange={(e) => setEditForm(f => ({ ...f, controlType: e.target.value, grade: '' }))}>
                  <option value="">Не изменять</option>
                  <option value="Зачет">Зачет</option>
                  <option value="Экзамен">Экзамен</option>
                  <option value="Зачет с оценкой">Зачет с оценкой</option>
                </select>
              </div>
              {editForm.controlType === 'Зачет' && (
                <div className="form-group"><label>Оценка</label>
                  <select value={editForm.grade} onChange={(e) => setEditForm(f => ({ ...f, grade: e.target.value }))}>
                    <option value="">Выберите...</option>
                    <option value="Зачтено">Зачтено</option>
                    <option value="Не зачтено">Не зачтено</option>
                  </select>
                </div>
              )}
              {(editForm.controlType === 'Экзамен' || editForm.controlType === 'Зачет с оценкой') && (
                <div className="form-group"><label>Оценка</label>
                  <select value={editForm.grade} onChange={(e) => setEditForm(f => ({ ...f, grade: e.target.value }))}>
                    <option value="">Выберите...</option>
                    <option value="Отлично">Отлично</option>
                    <option value="Хорошо">Хорошо</option>
                    <option value="Удовлетворительно">Удовлетворительно</option>
                    <option value="Неудовлетворительно">Неудовлетворительно</option>
                  </select>
                </div>
              )}
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
