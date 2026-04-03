import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { AlertTriangle, Search, Scale, Filter } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';
import StatCard from '../components/UI/StatCard';
import Modal from '../components/UI/Modal';

const groupModes = [
  { key: 'none', label: 'Без группировки' },
  { key: 'discipline', label: 'По дисциплинам' },
  { key: 'teacher', label: 'По преподавателям' },
  { key: 'student', label: 'По студентам' },
];

export default function MonitoringPage() {
  const { groups, disciplines, teachers, semesters, fetchMonitoring, addCommission, stats } = useData();
  const [results, setResults] = useState([]);
  const [groupFilter, setGroupFilter] = useState('');
  const [disciplineFilter, setDisciplineFilter] = useState('');
  const [semesterFilter, setSemesterFilter] = useState('');
  const [teacherFilter, setTeacherFilter] = useState('');
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [groupMode, setGroupMode] = useState('none');

  const [showCommModal, setShowCommModal] = useState(false);
  const [commTarget, setCommTarget] = useState(null);
  const [commDesc, setCommDesc] = useState('');
  const [commDate, setCommDate] = useState('');
  const [commMembers, setCommMembers] = useState([]);

  const loadData = async () => {
    setLoading(true);
    const filters = {};
    if (groupFilter) filters.group_id = groupFilter;
    if (disciplineFilter) filters.discipline_id = disciplineFilter;
    if (semesterFilter) filters.semester = semesterFilter;
    if (teacherFilter) filters.teacher_id = teacherFilter;
    const data = await fetchMonitoring(filters);
    setResults(data);
    setLoading(false);
  };

  useEffect(() => { loadData(); }, [groupFilter, disciplineFilter, semesterFilter, teacherFilter]);

  const filtered = results.filter(r => !search || r.student_fio?.toLowerCase().includes(search.toLowerCase()));

  const openCommission = (item) => {
    setCommTarget(item);
    setCommDesc(`Студент ${item.student_fio} имеет задолженность по дисциплине «${item.discipline_name}». Оценка: ${item.grade}.`);
    setCommDate('');
    setCommMembers([]);
    setShowCommModal(true);
  };

  const toggleMember = (tid) => {
    setCommMembers(prev => prev.includes(tid) ? prev.filter(id => id !== tid) : [...prev, tid]);
  };

  const handleCreateCommission = async (e) => {
    e.preventDefault();
    const members = commMembers.map(tid => ({
      teacher_id: tid,
      role: 'Член комиссии'
    }));
    await addCommission({
      performance_id: commTarget.performance_id,
      description: commDesc,
      meeting_date: commDate,
      members,
    });
    setShowCommModal(false);
    loadData();
  };

  // Grouping logic
  const getGrouped = () => {
    if (groupMode === 'discipline') {
      const map = {};
      filtered.forEach(r => { const k = r.discipline_name || 'Без дисциплины'; if (!map[k]) map[k] = []; map[k].push(r); });
      return map;
    }
    if (groupMode === 'teacher') {
      const map = {};
      filtered.forEach(r => { const k = r.teacher_fio || 'Не назначен'; if (!map[k]) map[k] = []; map[k].push(r); });
      return map;
    }
    if (groupMode === 'student') {
      const map = {};
      filtered.forEach(r => { const k = `${r.student_fio} (${r.group_name})`; if (!map[k]) map[k] = []; map[k].push(r); });
      return map;
    }
    return null;
  };

  const grouped = getGrouped();

  const renderTable = (rows) => (
    <table>
      <thead><tr>
        <th>Студент</th><th>Группа</th><th>Дисциплина</th><th>Оценка</th>
        <th>Задолженность</th><th>Комментарий</th><th>Действие</th>
      </tr></thead>
      <tbody>
        {rows.map((r, i) => (
          <motion.tr key={r.performance_id} initial={{ opacity: 0 }} animate={{ opacity: 1 }} transition={{ delay: i * 0.02 }}
            style={r.has_debt ? { background: 'rgba(220, 38, 38, 0.03)' } : {}}>
            <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{r.student_fio}</td>
            <td>{r.group_name}</td>
            <td>{r.discipline_name}</td>
            <td><StatusBadge status={r.grade} /></td>
            <td>{r.has_debt ? <span style={{ color: 'var(--danger)', fontWeight: 600, fontSize: '0.8rem' }}>Да</span> : <span style={{ color: 'var(--success)', fontSize: '0.8rem' }}>Нет</span>}</td>
            <td style={{ maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{r.comment || '—'}</td>
            <td>{r.has_commission ? <span style={{ fontSize: '0.8rem', color: 'var(--text-muted)' }}>На комиссии</span> : (
              <button className="btn btn-warning btn-sm" onClick={() => openCommission(r)}><Scale size={14} /> На комиссию</button>
            )}</td>
          </motion.tr>
        ))}
      </tbody>
    </table>
  );

  return (
    <div className="page">
      <div className="page-header"><h2 className="page-title">Мониторинг успеваемости</h2></div>

      <div className="stats-grid">
        <StatCard icon={<AlertTriangle size={24} />} label="Задолженностей" value={stats.total_debts || 0} color="danger" delay={0} />
        <StatCard icon={<Scale size={24} />} label="Комиссии в работе" value={stats.total_commissions || 0} color="warning" delay={0.1} />
        <StatCard icon={<Filter size={24} />} label="Найдено записей" value={filtered.length} color="info" delay={0.2} />
      </div>

      {filtered.length > 0 && (
        <div className="monitoring-alert"><AlertTriangle size={20} />
          <span>Обнаружено <strong>{filtered.length}</strong> записей с неудовлетворительной успеваемостью</span></div>
      )}

      <div className="filters-bar">
        <div className="search-input"><Search size={16} />
          <input placeholder="Поиск по ФИО..." value={search} onChange={(e) => setSearch(e.target.value)} /></div>
        <select value={groupFilter} onChange={(e) => setGroupFilter(e.target.value)}>
          <option value="">Все группы</option>
          {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
        </select>
        <select value={disciplineFilter} onChange={(e) => setDisciplineFilter(e.target.value)}>
          <option value="">Все дисциплины</option>
          {disciplines.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
        </select>
        <select value={teacherFilter} onChange={(e) => setTeacherFilter(e.target.value)}>
          <option value="">Все преподаватели</option>
          {teachers.map(t => <option key={t.id} value={t.id}>{t.fio}</option>)}
        </select>
        <select value={semesterFilter} onChange={(e) => setSemesterFilter(e.target.value)}>
          <option value="">Все семестры</option>
          {semesters.map(s => <option key={s} value={s}>{s} семестр</option>)}
        </select>
      </div>

      <div style={{ display: 'flex', gap: '6px', marginBottom: '16px', flexWrap: 'wrap' }}>
        {groupModes.map(m => (
          <button key={m.key} className={`btn btn-sm ${groupMode === m.key ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setGroupMode(m.key)} style={{ fontSize: '0.8rem', padding: '6px 14px' }}>{m.label}</button>
        ))}
      </div>

      {grouped ? (
        Object.entries(grouped).map(([title, rows]) => (
          <motion.div key={title} className="table-container" style={{ marginBottom: '20px' }}
            initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }}>
            <div style={{ padding: '10px 16px', fontWeight: 700, fontSize: '0.95rem', color: 'var(--accent)', borderBottom: '1px solid var(--border)', background: 'var(--bg-secondary)' }}>
              {title} <span style={{ fontWeight: 400, color: 'var(--text-muted)', fontSize: '0.8rem' }}>({rows.length} зап.)</span>
            </div>
            {renderTable(rows)}
          </motion.div>
        ))
      ) : (
        <motion.div className="table-container" layout>
          {renderTable(filtered)}
          {!loading && filtered.length === 0 && (<div className="empty-state"><AlertTriangle size={48} /><p>Проблемных студентов не обнаружено</p></div>)}
        </motion.div>
      )}

      <Modal isOpen={showCommModal} onClose={() => setShowCommModal(false)} title="Направить на комиссию" wide>
        {commTarget && (
          <form onSubmit={handleCreateCommission}>
            <div className="detail-grid" style={{ marginBottom: '16px' }}>
              <div className="detail-item"><span className="detail-label">Студент</span><span className="detail-value">{commTarget.student_fio}</span></div>
              <div className="detail-item"><span className="detail-label">Группа</span><span className="detail-value">{commTarget.group_name}</span></div>
              <div className="detail-item"><span className="detail-label">Дисциплина</span><span className="detail-value">{commTarget.discipline_name}</span></div>
              <div className="detail-item"><span className="detail-label">Оценка</span><span className="detail-value"><StatusBadge status={commTarget.grade} /></span></div>
            </div>
            <div className="form-group"><label>Дата заседания комиссии</label>
              <input type="date" value={commDate} onChange={(e) => setCommDate(e.target.value)} /></div>
            <div className="form-group"><label>Члены комиссии</label>
              <div style={{ display: 'flex', flexDirection: 'column', gap: '6px', padding: '8px', background: 'var(--bg-tertiary)', borderRadius: '8px', maxHeight: '160px', overflowY: 'auto' }}>
                {teachers.map(t => (
                  <label key={t.id} style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '0.875rem' }}>
                    <input type="checkbox" checked={commMembers.includes(t.id)} onChange={() => toggleMember(t.id)} />
                    {t.fio} <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>({t.position})</span>
                  </label>
                ))}
              </div></div>
            <div className="form-group"><label>Описание / Основание</label>
              <textarea rows={3} required value={commDesc} onChange={(e) => setCommDesc(e.target.value)} /></div>
            <div className="modal-actions">
              <button type="button" className="btn btn-secondary" onClick={() => setShowCommModal(false)}>Отмена</button>
              <button type="submit" className="btn btn-warning"><Scale size={16} /> Направить</button>
            </div>
          </form>
        )}
      </Modal>
    </div>
  );
}
