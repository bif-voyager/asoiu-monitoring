import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertTriangle, Search, Scale } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';
import StatCard from '../components/UI/StatCard';
import Modal from '../components/UI/Modal';

export default function MonitoringPage() {
  const { groups, disciplines, semesters, fetchMonitoring, addCommission, stats } = useData();
  const [results, setResults] = useState([]);
  const [groupFilter, setGroupFilter] = useState('');
  const [disciplineFilter, setDisciplineFilter] = useState('');
  const [semesterFilter, setSemesterFilter] = useState('');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);

  const [showCommModal, setShowCommModal] = useState(false);
  const [commTarget, setCommTarget] = useState(null);
  const [commDesc, setCommDesc] = useState('');

  const loadData = async () => {
    setLoading(true);
    const filters = {};
    if (groupFilter) filters.group = groupFilter;
    if (disciplineFilter) filters.discipline_id = disciplineFilter;
    if (semesterFilter) filters.semester = semesterFilter;
    const data = await fetchMonitoring(filters);
    setResults(data);
    setLoading(false);
  };

  useEffect(() => { loadData(); }, [groupFilter, disciplineFilter, semesterFilter]);

  const filtered = results.filter(r => {
    return !search || r.student_fio?.toLowerCase().includes(search.toLowerCase());
  });

  const openCommission = (item) => {
    setCommTarget(item);
    setCommDesc(`Студент ${item.student_fio} имеет задолженность по дисциплине «${item.discipline_name}». Оценка: ${item.grade}.`);
    setShowCommModal(true);
  };

  const handleCreateCommission = async (e) => {
    e.preventDefault();
    await addCommission({
      student_id: commTarget.student_id,
      performance_id: commTarget.performance_id,
      description: commDesc,
    });
    setShowCommModal(false);
    loadData();
  };

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">Мониторинг успеваемости</h2>
      </div>

      <div className="stats-grid">
        <StatCard icon={<AlertTriangle size={24} />} label="Задолженностей" value={stats.total_debts || 0} color="danger" delay={0} />
        <StatCard icon={<Scale size={24} />} label="Активных комиссий" value={stats.total_commissions || 0} color="warning" delay={0.1} />
      </div>

      {filtered.length > 0 && (
        <div className="monitoring-alert">
          <AlertTriangle size={20} />
          <span>Обнаружено <strong>{filtered.length}</strong> записей с неудовлетворительной успеваемостью</span>
        </div>
      )}

      <div className="filters-bar">
        <div className="search-input">
          <Search size={16} />
          <input
            placeholder="Поиск по ФИО..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <select value={groupFilter} onChange={(e) => setGroupFilter(e.target.value)}>
          <option value="">Все группы</option>
          {groups.map(g => <option key={g} value={g}>{g}</option>)}
        </select>
        <select value={disciplineFilter} onChange={(e) => setDisciplineFilter(e.target.value)}>
          <option value="">Все дисциплины</option>
          {disciplines.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
        </select>
        <select value={semesterFilter} onChange={(e) => setSemesterFilter(e.target.value)}>
          <option value="">Все семестры</option>
          {semesters.map(s => <option key={s} value={s}>{s} семестр</option>)}
        </select>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead>
            <tr>
              <th>Студент</th>
              <th>Группа</th>
              <th>Курс</th>
              <th>Дисциплина</th>
              <th>Оценка</th>
              <th>Ср. балл</th>
              <th>Задолж.</th>
              <th>Комментарий</th>
              <th>Действие</th>
            </tr>
          </thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((r, i) => (
                <motion.tr
                  key={`${r.performance_id}`}
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0 }}
                  transition={{ delay: i * 0.02 }}
                  style={r.has_debt ? { background: 'rgba(220, 38, 38, 0.03)' } : {}}
                >
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{r.student_fio}</td>
                  <td>{r.group_name}</td>
                  <td>{r.course}</td>
                  <td>{r.discipline_name}</td>
                  <td><StatusBadge status={r.grade} /></td>
                  <td style={{ fontWeight: 500 }}>{r.avg_grade}</td>
                  <td>
                    {r.has_debt ? (
                      <span style={{ color: 'var(--danger)', fontWeight: 600, fontSize: '0.8rem' }}>Да</span>
                    ) : (
                      <span style={{ color: 'var(--success)', fontSize: '0.8rem' }}>Нет</span>
                    )}
                  </td>
                  <td style={{ maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {r.comment || '—'}
                  </td>
                  <td>
                    {r.has_commission ? (
                      <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>На комиссии</span>
                    ) : (
                      <button
                        className="btn btn-warning btn-sm"
                        onClick={() => openCommission(r)}
                      >
                        <Scale size={14} /> На комиссию
                      </button>
                    )}
                  </td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
        {!loading && filtered.length === 0 && (
          <div className="empty-state">
            <AlertTriangle size={48} />
            <p>Проблемных студентов не обнаружено</p>
          </div>
        )}
      </motion.div>

      <Modal isOpen={showCommModal} onClose={() => setShowCommModal(false)} title="Направить на комиссию">
        {commTarget && (
          <form onSubmit={handleCreateCommission}>
            <div className="detail-grid" style={{ marginBottom: '16px' }}>
              <div className="detail-item">
                <span className="detail-label">Студент</span>
                <span className="detail-value">{commTarget.student_fio}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Группа</span>
                <span className="detail-value">{commTarget.group_name}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Дисциплина</span>
                <span className="detail-value">{commTarget.discipline_name}</span>
              </div>
              <div className="detail-item">
                <span className="detail-label">Оценка</span>
                <span className="detail-value"><StatusBadge status={commTarget.grade} /></span>
              </div>
            </div>
            <div className="form-group">
              <label>Описание / Основание</label>
              <textarea
                rows={3}
                required
                value={commDesc}
                onChange={(e) => setCommDesc(e.target.value)}
              />
            </div>
            <div className="modal-actions">
              <button type="button" className="btn btn-secondary" onClick={() => setShowCommModal(false)}>Отмена</button>
              <button type="submit" className="btn btn-warning">
                <Scale size={16} /> Направить
              </button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  );
}
