import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Search, Edit2, Trash2, BookOpen } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';
import Modal from '../components/UI/Modal';

const emptyForm = { name: '', semester: 1, control_type: 'Экзамен', hours: 72, teacher_id: '' };

export default function DisciplinesPage() {
  const { disciplines, teachers, semesters, addDiscipline, updateDiscipline, deleteDiscipline } = useData();
  const [search, setSearch] = useState('');
  const [semesterFilter, setSemesterFilter] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({ ...emptyForm });

  const filtered = disciplines.filter(d => {
    const matchSearch = !search || d.name.toLowerCase().includes(search.toLowerCase());
    const matchSemester = !semesterFilter || d.semester === Number(semesterFilter);
    return matchSearch && matchSemester;
  });

  const openCreate = () => {
    setForm({ ...emptyForm, teacher_id: teachers[0]?.id || '' });
    setEditId(null);
    setShowModal(true);
  };

  const openEdit = (d) => {
    setForm({ name: d.name, semester: d.semester, control_type: d.control_type, hours: d.hours, teacher_id: d.teacher_id });
    setEditId(d.id);
    setShowModal(true);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (editId) {
      updateDiscipline(editId, form);
    } else {
      addDiscipline(form);
    }
    setShowModal(false);
  };

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">Дисциплины</h2>
        <button className="btn btn-primary" onClick={openCreate}>
          <Plus size={18} /> Добавить дисциплину
        </button>
      </div>

      <div className="filters-bar">
        <div className="search-input">
          <Search size={16} />
          <input
            placeholder="Поиск по названию..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <select value={semesterFilter} onChange={(e) => setSemesterFilter(e.target.value)}>
          <option value="">Все семестры</option>
          {semesters.map(s => <option key={s} value={s}>{s} семестр</option>)}
        </select>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead>
            <tr>
              <th>Название</th>
              <th>Семестр</th>
              <th>Контроль</th>
              <th>Часы</th>
              <th>Преподаватель</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((d, i) => (
                <motion.tr
                  key={d.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ delay: i * 0.03 }}
                >
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{d.name}</td>
                  <td>{d.semester}</td>
                  <td><StatusBadge status={d.control_type} /></td>
                  <td>{d.hours}</td>
                  <td>{d.teacher_fio || '—'}</td>
                  <td>
                    <div style={{ display: 'flex', gap: '4px', justifyContent: 'flex-end' }}>
                      <button className="btn-icon" onClick={() => openEdit(d)}><Edit2 size={16} /></button>
                      <button className="btn-delete" onClick={() => { if (confirm(`Удалить дисциплину "${d.name}"?`)) deleteDiscipline(d.id); }}>
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
        {filtered.length === 0 && (
          <div className="empty-state">
            <BookOpen size={48} />
            <p>Дисциплины не найдены</p>
          </div>
        )}
      </motion.div>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title={editId ? 'Редактирование дисциплины' : 'Новая дисциплина'}>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Название</label>
            <input required value={form.name} onChange={(e) => setForm(f => ({ ...f, name: e.target.value }))} placeholder="Базы данных" />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>Семестр</label>
              <select value={form.semester} onChange={(e) => setForm(f => ({ ...f, semester: Number(e.target.value) }))}>
                {[1,2,3,4,5,6,7,8].map(s => <option key={s} value={s}>{s}</option>)}
              </select>
            </div>
            <div className="form-group">
              <label>Тип контроля</label>
              <select value={form.control_type} onChange={(e) => setForm(f => ({ ...f, control_type: e.target.value }))}>
                <option>Экзамен</option>
                <option>Зачёт</option>
                <option>Курсовая работа</option>
              </select>
            </div>
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>Количество часов</label>
              <input type="number" min="1" required value={form.hours} onChange={(e) => setForm(f => ({ ...f, hours: Number(e.target.value) }))} />
            </div>
            <div className="form-group">
              <label>Преподаватель</label>
              <select required value={form.teacher_id} onChange={(e) => setForm(f => ({ ...f, teacher_id: Number(e.target.value) }))}>
                <option value="">Выберите...</option>
                {teachers.map(t => <option key={t.id} value={t.id}>{t.fio}</option>)}
              </select>
            </div>
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
