import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Search, Edit2, Trash2, Users } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';
import StatCard from '../components/UI/StatCard';
import Modal from '../components/UI/Modal';

const emptyForm = { fio: '', record_book_number: '', group_id: '', status: 'Активен' };

export default function StudentsPage() {
  const { students, groups, addStudent, updateStudent, deleteStudent } = useData();
  const [search, setSearch] = useState('');
  const [groupFilter, setGroupFilter] = useState('');
  const [statusFilter, setStatusFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({ ...emptyForm });

  const filtered = students.filter(s => {
    const matchSearch = !search ||
      s.fio.toLowerCase().includes(search.toLowerCase()) ||
      s.record_book_number.toLowerCase().includes(search.toLowerCase());
    const matchGroup = !groupFilter || s.group_id === Number(groupFilter);
    const matchStatus = !statusFilter || s.status === statusFilter;
    return matchSearch && matchGroup && matchStatus;
  });

  const openCreate = () => { setForm({ ...emptyForm }); setEditId(null); setShowModal(true); };
  const openEdit = (s) => {
    setForm({ fio: s.fio, record_book_number: s.record_book_number, group_id: s.group_id, status: s.status });
    setEditId(s.id); setShowModal(true);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (editId) { updateStudent(editId, form); } else { addStudent(form); }
    setShowModal(false);
  };

  const statuses = [...new Set(students.map(s => s.status))];
  const selectedGroup = groups.find(g => g.id === Number(form.group_id));

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">Студенты</h2>
        <button className="btn btn-primary" onClick={openCreate}><Plus size={18} /> Добавить студента</button>
      </div>

      <div className="stats-grid">
        <StatCard icon={<Users size={24} />} label="Всего студентов" value={students.length} color="accent" delay={0} />
        <StatCard icon={<Users size={24} />} label="Активных" value={students.filter(s => s.status === 'Активен').length} color="success" delay={0.1} />
        <StatCard icon={<Users size={24} />} label="Групп" value={groups.length} color="info" delay={0.2} />
      </div>

      <div className="filters-bar">
        <div className="search-input">
          <Search size={16} />
          <input placeholder="Поиск по ФИО или зачётке..." value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
        <select value={groupFilter} onChange={(e) => setGroupFilter(e.target.value)}>
          <option value="">Все группы</option>
          {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
        </select>
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}>
          <option value="">Все статусы</option>
          {statuses.map(s => <option key={s} value={s}>{s}</option>)}
        </select>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead>
            <tr>
              <th>ФИО</th><th>№ Зачётки</th><th>Группа</th><th>Курс</th>
              <th>Форма</th><th>Направление</th><th>Статус</th><th></th>
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((s, i) => (
                <motion.tr key={s.id} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} transition={{ delay: i * 0.02 }}>
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{s.fio}</td>
                  <td>{s.record_book_number}</td>
                  <td>{s.group_name}</td>
                  <td>{s.course}</td>
                  <td>{s.study_form}</td>
                  <td>{s.training_direction}</td>
                  <td><StatusBadge status={s.status} /></td>
                  <td>
                    <div style={{ display: 'flex', gap: '4px', justifyContent: 'flex-end' }}>
                      <button className="btn-icon" onClick={() => openEdit(s)}><Edit2 size={16} /></button>
                      <button className="btn-delete" onClick={() => { if (confirm(`Удалить студента ${s.fio}?`)) deleteStudent(s.id); }}><Trash2 size={16} /></button>
                    </div>
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
        {filtered.length === 0 && (<div className="empty-state"><Users size={48} /><p>Студенты не найдены</p></div>)}
      </motion.div>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title={editId ? 'Редактирование студента' : 'Новый студент'} wide>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>ФИО</label>
            <input required value={form.fio} onChange={(e) => setForm(f => ({ ...f, fio: e.target.value }))} placeholder="Фамилия Имя Отчество" />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>Номер зачётной книжки</label>
              <input required value={form.record_book_number} onChange={(e) => setForm(f => ({ ...f, record_book_number: e.target.value }))} placeholder="2024-001" />
            </div>
            <div className="form-group">
              <label>Группа</label>
              <select required value={form.group_id} onChange={(e) => setForm(f => ({ ...f, group_id: Number(e.target.value) }))}>
                <option value="">Выберите группу...</option>
                {groups.map(g => <option key={g.id} value={g.id}>{g.name} — {g.training_direction} ({g.study_form})</option>)}
              </select>
            </div>
          </div>
          {selectedGroup && (
            <div style={{ padding: '8px 12px', background: 'var(--bg-tertiary)', borderRadius: '8px', marginBottom: '12px', fontSize: '0.85rem', color: 'var(--text-secondary)' }}>
              <strong>Курс:</strong> {selectedGroup.course} · <strong>Направление:</strong> {selectedGroup.training_direction} · <strong>Форма:</strong> {selectedGroup.study_form}
            </div>
          )}
          <div className="form-group">
            <label>Статус</label>
            <select value={form.status} onChange={(e) => setForm(f => ({ ...f, status: e.target.value }))}>
              <option>Активен</option><option>Отчислен</option><option>Академический отпуск</option>
            </select>
          </div>
          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Отмена</button>
            <button type="submit" className="btn btn-primary">{editId ? 'Сохранить' : 'Создать'}</button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
