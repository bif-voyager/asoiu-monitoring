import { createContext, useContext, useState, useEffect } from 'react';

const DataContext = createContext(null);
const API = 'http://localhost:3000/api';

export function DataProvider({ children }) {
  const [students, setStudents] = useState([]);
  const [teachers, setTeachers] = useState([]);
  const [disciplines, setDisciplines] = useState([]);
  const [performance, setPerformance] = useState([]);
  const [commissions, setCommissions] = useState([]);
  const [groups, setGroups] = useState([]);
  const [semesters, setSemesters] = useState([]);
  const [stats, setStats] = useState({});

  const fetchAll = async () => {
    try {
      const endpoints = [
        'students', 'teachers', 'disciplines', 'performance', 'commissions',
        'groups', 'semesters', 'stats'
      ];
      const responses = await Promise.all(endpoints.map(e => fetch(`${API}/${e}`)));
      const data = await Promise.all(responses.map(r => r.json()));
      setStudents(data[0]); setTeachers(data[1]); setDisciplines(data[2]);
      setPerformance(data[3]); setCommissions(data[4]); setGroups(data[5]);
      setSemesters(data[6]); setStats(data[7]);
    } catch (e) { console.error("Failed to fetch data", e); }
  };

  useEffect(() => { fetchAll(); }, []);

  const apiPost = async (endpoint, data) => {
    const res = await fetch(`${API}/${endpoint}`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data)
    });
    return res.json();
  };
  const apiPut = async (endpoint, id, data) => {
    const res = await fetch(`${API}/${endpoint}/${id}`, {
      method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(data)
    });
    return res.json();
  };
  const apiDelete = async (endpoint, id) => { await fetch(`${API}/${endpoint}/${id}`, { method: 'DELETE' }); };

  const addStudent = async (s) => { await apiPost('students', s); fetchAll(); };
  const updateStudent = async (id, s) => { await apiPut('students', id, s); fetchAll(); };
  const deleteStudent = async (id) => { await apiDelete('students', id); fetchAll(); };

  const addTeacher = async (t) => { await apiPost('teachers', t); fetchAll(); };
  const updateTeacher = async (id, t) => { await apiPut('teachers', id, t); fetchAll(); };
  const deleteTeacher = async (id) => { await apiDelete('teachers', id); fetchAll(); };

  const addDiscipline = async (d) => { await apiPost('disciplines', d); fetchAll(); };
  const updateDiscipline = async (id, d) => { await apiPut('disciplines', id, d); fetchAll(); };
  const deleteDiscipline = async (id) => { await apiDelete('disciplines', id); fetchAll(); };

  const addPerformance = async (p) => { await apiPost('performance', p); fetchAll(); };
  const updatePerformance = async (id, p) => { await apiPut('performance', id, p); fetchAll(); };
  const deletePerformance = async (id) => { await apiDelete('performance', id); fetchAll(); };
  const batchPerformance = async (records) => {
    await fetch(`${API}/performance/batch`, {
      method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ records })
    });
    fetchAll();
  };

  const createSheet = async (data) => { return apiPost('sheets', data); };

  const addCommission = async (c) => { await apiPost('commissions', c); fetchAll(); };
  const updateCommission = async (id, c) => { await apiPut('commissions', id, c); fetchAll(); };

  const fetchMonitoring = async (filters = {}) => {
    const params = new URLSearchParams();
    if (filters.group_id) params.set('group_id', filters.group_id);
    if (filters.discipline_id) params.set('discipline_id', filters.discipline_id);
    if (filters.semester) params.set('semester', filters.semester);
    if (filters.teacher_id) params.set('teacher_id', filters.teacher_id);
    const res = await fetch(`${API}/monitoring?${params}`);
    return res.json();
  };

  const fetchStudentsByGroup = async (groupId) => {
    const res = await fetch(`${API}/students-by-group?group_id=${groupId}`);
    return res.json();
  };

  const getStudent = (id) => students.find(s => s.id === id);
  const getTeacher = (id) => teachers.find(t => t.id === id);
  const getDiscipline = (id) => disciplines.find(d => d.id === id);

  return (
    <DataContext.Provider value={{
      students, teachers, disciplines, performance, commissions, groups, semesters, stats,
      addStudent, updateStudent, deleteStudent,
      addTeacher, updateTeacher, deleteTeacher,
      addDiscipline, updateDiscipline, deleteDiscipline,
      addPerformance, updatePerformance, deletePerformance, batchPerformance,
      createSheet, addCommission, updateCommission,
      fetchMonitoring, fetchStudentsByGroup,
      getStudent, getTeacher, getDiscipline, fetchAll
    }}>
      {children}
    </DataContext.Provider>
  );
}

export const useData = () => useContext(DataContext);
