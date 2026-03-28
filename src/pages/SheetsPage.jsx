import { useState } from 'react';
import { motion } from 'framer-motion';
import { Save, ClipboardList, Users } from 'lucide-react';
import { useData } from '../context/DataContext';

const gradeOptions = ['', 'Отлично', 'Хорошо', 'Удовлетворительно', 'Неудовлетворительно', 'Зачтено', 'Не зачтено'];

export default function SheetsPage() {
  const { groups, disciplines, teachers, fetchStudentsByGroup, batchPerformance } = useData();
  const [selectedGroup, setSelectedGroup] = useState('');
  const [selectedDiscipline, setSelectedDiscipline] = useState('');
  const [sheetNumber, setSheetNumber] = useState('');
  const [studentRows, setStudentRows] = useState([]);
  const [loaded, setLoaded] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);

  const discipline = disciplines.find(d => d.id === Number(selectedDiscipline));

  const loadStudents = async () => {
    if (!selectedGroup || !selectedDiscipline || !sheetNumber) {
      alert('Заполните все поля: группа, дисциплина, номер ведомости');
      return;
    }
    const students = await fetchStudentsByGroup(selectedGroup);
    setStudentRows(students.map(s => ({
      student_id: s.id,
      fio: s.fio,
      record_book: s.record_book_number,
      grade: '',
      comment: '',
    })));
    setLoaded(true);
    setSaved(false);
  };

  const updateRow = (idx, field, value) => {
    setStudentRows(prev => {
      const newRows = [...prev];
      newRows[idx] = { ...newRows[idx], [field]: value };
      return newRows;
    });
    setSaved(false);
  };

  const handleSave = async () => {
    const records = studentRows
      .filter(r => r.grade)
      .map(r => ({
        student_id: r.student_id,
        discipline_id: Number(selectedDiscipline),
        teacher_id: discipline?.teacher_id || null,
        sheet_number: sheetNumber,
        grade: r.grade,
        comment: r.comment,
      }));

    if (records.length === 0) {
      alert('Поставьте хотя бы одну оценку');
      return;
    }

    setSaving(true);
    await batchPerformance(records);
    setSaving(false);
    setSaved(true);
  };

  return (
    <div className="page">
      <div className="page-header">
        <h2 className="page-title">Заполнение ведомости</h2>
      </div>

      <div className="sheet-setup">
        <div className="form-group">
          <label>Группа</label>
          <select value={selectedGroup} onChange={(e) => { setSelectedGroup(e.target.value); setLoaded(false); }}>
            <option value="">Выберите группу...</option>
            {groups.map(g => <option key={g} value={g}>{g}</option>)}
          </select>
        </div>
        <div className="form-group">
          <label>Дисциплина</label>
          <select value={selectedDiscipline} onChange={(e) => { setSelectedDiscipline(e.target.value); setLoaded(false); }}>
            <option value="">Выберите дисциплину...</option>
            {disciplines.map(d => <option key={d.id} value={d.id}>{d.name} ({d.semester} сем.)</option>)}
          </select>
        </div>
        <div className="form-group">
          <label>Номер ведомости</label>
          <input
            value={sheetNumber}
            onChange={(e) => { setSheetNumber(e.target.value); setLoaded(false); }}
            placeholder="ВД-001"
          />
        </div>
        <button className="btn btn-primary" onClick={loadStudents} style={{ marginTop: '22px' }}>
          <Users size={18} /> Загрузить студентов
        </button>
      </div>

      {loaded && studentRows.length > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
        >
          {discipline && (
            <div style={{ marginBottom: '16px', display: 'flex', gap: '24px', fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
              <span><strong>Преподаватель:</strong> {discipline.teacher_fio || '—'}</span>
              <span><strong>Тип контроля:</strong> {discipline.control_type}</span>
              <span><strong>Семестр:</strong> {discipline.semester}</span>
            </div>
          )}

          <div className="sheet-table-container">
            <table>
              <thead>
                <tr>
                  <th style={{ width: '40px' }}>№</th>
                  <th>ФИО студента</th>
                  <th>№ Зачётки</th>
                  <th style={{ width: '220px' }}>Оценка</th>
                  <th>Комментарий</th>
                </tr>
              </thead>
              <tbody>
                {studentRows.map((row, i) => (
                  <tr key={row.student_id}>
                    <td style={{ color: 'var(--text-muted)' }}>{i + 1}</td>
                    <td style={{ fontWeight: 500, color: 'var(--text-primary)' }}>{row.fio}</td>
                    <td>{row.record_book}</td>
                    <td>
                      <select
                        value={row.grade}
                        onChange={(e) => updateRow(i, 'grade', e.target.value)}
                        style={{ minWidth: '180px' }}
                      >
                        {gradeOptions.map(g => (
                          <option key={g} value={g}>{g || '— Не выбрано —'}</option>
                        ))}
                      </select>
                    </td>
                    <td>
                      <input
                        value={row.comment}
                        onChange={(e) => updateRow(i, 'comment', e.target.value)}
                        placeholder="Комментарий..."
                        style={{ minWidth: '150px' }}
                      />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '12px', marginTop: '20px', alignItems: 'center' }}>
            {saved && (
              <motion.span
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                style={{ color: 'var(--success)', fontSize: '0.875rem', fontWeight: 500 }}
              >
                ✓ Ведомость сохранена
              </motion.span>
            )}
            <button className="btn btn-success" onClick={handleSave} disabled={saving}>
              <Save size={18} /> {saving ? 'Сохранение...' : 'Сохранить ведомость'}
            </button>
          </div>
        </motion.div>
      )}

      {loaded && studentRows.length === 0 && (
        <div className="empty-state">
          <ClipboardList size={48} />
          <p>В выбранной группе нет активных студентов</p>
        </div>
      )}

      {!loaded && (
        <div className="empty-state">
          <ClipboardList size={48} />
          <p>Выберите группу, дисциплину и номер ведомости, затем нажмите «Загрузить студентов»</p>
        </div>
      )}
    </div>
  );
}
