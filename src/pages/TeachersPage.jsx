import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Plus, Search, Edit2, Trash2, GraduationCap } from 'lucide-react';
import { useData } from '../context/DataContext';
import Modal from '../components/UI/Modal';

const emptyForm = { fio: '', position: '', academic_degree: '', login: '', password: '' };

export default function TeachersPage() {
  const { teachers, addTeacher, updateTeacher, deleteTeacher } = useData();
  const [search, setSearch] = useState('');
  const [showModal, setShowModal] = useState(false);
  const [editId, setEditId] = useState(null);
  const [form, setForm] = useState({ ...emptyForm });

  const filtered = teachers.filter(t => {
    return !search ||
      t.fio.toLowerCase().includes(search.toLowerCase()) ||
      t.position.toLowerCase().includes(search.toLowerCase());
  });

  const openCreate = () => {
    setForm({ ...emptyForm });
    setEditId(null);
    setShowModal(true);
  };

  const openEdit = (t) => {
    setForm({ fio: t.fio, position: t.position, academic_degree: t.academic_degree, login: t.login, password: '' });
    setEditId(t.id);
    setShowModal(true);
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (editId) {
      updateTeacher(editId, form);
    } else {
      addTeacher(form);
    }
    setShowModal(false);
  };

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">Преподаватели</h2>
      </div>

      <div className="filters-bar">
        <div className="search-input">
          <Search size={16} />
          <input
            placeholder="Поиск по ФИО или должности..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead>
            <tr>
              <th>ФИО</th>
              <th>Должность</th>
              <th>Учёная степень</th>
              <th>Логин</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((t, i) => (
                <motion.tr
                  key={t.id}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ delay: i * 0.03 }}
                >
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{t.fio}</td>
                  <td>{t.position}</td>
                  <td>{t.academic_degree}</td>
                  <td><code style={{ background: 'var(--bg-tertiary)', padding: '2px 8px', borderRadius: '4px', fontSize: '0.8rem' }}>{t.login}</code></td>
                  <td>
                    <div style={{ display: 'flex', gap: '4px', justifyContent: 'flex-end' }}>
                      <button className="btn-icon" onClick={() => openEdit(t)}><Edit2 size={16} /></button>
                      <button className="btn-delete" onClick={() => { if (confirm(`Удалить преподавателя ${t.fio}?`)) deleteTeacher(t.id); }}>
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
            <GraduationCap size={48} />
            <p>Преподаватели не найдены</p>
          </div>
        )}
      </motion.div>

      <Modal isOpen={showModal} onClose={() => setShowModal(false)} title={editId ? 'Редактирование преподавателя' : 'Новый преподаватель'}>
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>ФИО</label>
            <input required value={form.fio} onChange={(e) => setForm(f => ({ ...f, fio: e.target.value }))} placeholder="Фамилия Имя Отчество" />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>Должность</label>
              <select value={form.position} onChange={(e) => setForm(f => ({ ...f, position: e.target.value }))}>
                <option value="">Выберите...</option>
                <option>Профессор</option>
                <option>Доцент</option>
                <option>Старший преподаватель</option>
                <option>Преподаватель</option>
                <option>Ассистент</option>
              </select>
            </div>
            <div className="form-group">
              <label>Учёная степень</label>
              <select value={form.academic_degree} onChange={(e) => setForm(f => ({ ...f, academic_degree: e.target.value }))}>
                <option value="">Нет</option>
                <option>Кандидат наук</option>
                <option>Доктор наук</option>
              </select>
            </div>
          </div>
          {!editId && (
            <div className="form-row">
              <div className="form-group">
                <label>Логин</label>
                <input required value={form.login} onChange={(e) => setForm(f => ({ ...f, login: e.target.value }))} placeholder="login" />
              </div>
              <div className="form-group">
                <label>Пароль</label>
                <input type="password" value={form.password} onChange={(e) => setForm(f => ({ ...f, password: e.target.value }))} placeholder="По умолчанию: 123456" />
              </div>
            </div>
          )}
          <div className="modal-actions">
            <button type="button" className="btn btn-secondary" onClick={() => setShowModal(false)}>Отмена</button>
            <button type="submit" className="btn btn-primary">{editId ? 'Сохранить' : 'Создать'}</button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
