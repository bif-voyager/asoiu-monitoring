import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, BarChart3 } from 'lucide-react';
import { useData } from '../context/DataContext';
import StatusBadge from '../components/UI/StatusBadge';

export default function PerformancePage() {
  const { performance, groups, disciplines, semesters } = useData();
  const [search, setSearch] = useState('');
  const [groupFilter, setGroupFilter] = useState('');
  const [disciplineFilter, setDisciplineFilter] = useState('');
  const [semesterFilter, setSemesterFilter] = useState('');
  const [debtFilter, setDebtFilter] = useState('');

  const filtered = performance.filter(p => {
    const matchSearch = !search ||
      p.student_fio?.toLowerCase().includes(search.toLowerCase()) ||
      p.sheet_number?.toLowerCase().includes(search.toLowerCase());
    const matchGroup = !groupFilter || p.group_name === groups.find(g => g.id === Number(groupFilter))?.name;
    const matchDisc = !disciplineFilter || p.discipline_name === disciplines.find(d => d.id === Number(disciplineFilter))?.name;
    const matchSem = !semesterFilter || p.semester === Number(semesterFilter);
    const matchDebt = !debtFilter || (debtFilter === 'debt' && p.has_debt) || (debtFilter === 'no_debt' && !p.has_debt);
    return matchSearch && matchGroup && matchDisc && matchSem && matchDebt;
  });

  return (
    <div className="page">
      <div className="page-header"><h2 className="page-title">Успеваемость</h2></div>
      <div className="filters-bar">
        <div className="search-input"><Search size={16} />
          <input placeholder="Поиск по ФИО или номеру ведомости..." value={search} onChange={(e) => setSearch(e.target.value)} /></div>
        <select value={groupFilter} onChange={(e) => setGroupFilter(e.target.value)}>
          <option value="">Все группы</option>
          {groups.map(g => <option key={g.id} value={g.id}>{g.name}</option>)}
        </select>
        <select value={disciplineFilter} onChange={(e) => setDisciplineFilter(e.target.value)}>
          <option value="">Все дисциплины</option>
          {disciplines.map(d => <option key={d.id} value={d.id}>{d.name}</option>)}
        </select>
        <select value={semesterFilter} onChange={(e) => setSemesterFilter(e.target.value)}>
          <option value="">Все семестры</option>
          {semesters.map(s => <option key={s} value={s}>{s} семестр</option>)}
        </select>
        <select value={debtFilter} onChange={(e) => setDebtFilter(e.target.value)}>
          <option value="">Все записи</option>
          <option value="debt">С задолженностью</option>
          <option value="no_debt">Без задолженности</option>
        </select>
      </div>

      <motion.div className="table-container" layout>
        <table>
          <thead><tr>
            <th>Студент</th><th>Группа</th><th>Дисциплина</th><th>Семестр</th>
            <th>Ведомость</th><th>Оценка</th><th>Статус</th><th>Задолженность</th><th>Комментарий</th>
          </tr></thead>
          <tbody>
            <AnimatePresence>
              {filtered.map((p, i) => (
                <motion.tr key={p.id} initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} exit={{ opacity: 0 }} transition={{ delay: i * 0.02 }}>
                  <td style={{ fontWeight: 600, color: 'var(--text-primary)' }}>{p.student_fio}</td>
                  <td>{p.group_name}</td>
                  <td>{p.discipline_name}</td>
                  <td>{p.semester}</td>
                  <td><code style={{ background: 'var(--bg-tertiary)', padding: '2px 8px', borderRadius: '4px', fontSize: '0.8rem' }}>{p.sheet_number}</code></td>
                  <td><StatusBadge status={p.grade} /></td>
                  <td><StatusBadge status={p.status} /></td>
                  <td>{p.has_debt ? <span style={{ color: 'var(--danger)', fontWeight: 600, fontSize: '0.8rem' }}>Да</span> : <span style={{ color: 'var(--success)', fontSize: '0.8rem' }}>Нет</span>}</td>
                  <td style={{ maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{p.comment || '—'}</td>
                </motion.tr>
              ))}
            </AnimatePresence>
          </tbody>
        </table>
        {filtered.length === 0 && (<div className="empty-state"><BarChart3 size={48} /><p>Записи об успеваемости не найдены</p></div>)}
      </motion.div>
    </div>
  );
}
