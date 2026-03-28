import { Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from './context/AuthContext';
import AppLayout from './components/Layout/AppLayout';
import LoginPage from './pages/LoginPage';
import StudentsPage from './pages/StudentsPage';
import TeachersPage from './pages/TeachersPage';
import DisciplinesPage from './pages/DisciplinesPage';
import SheetsPage from './pages/SheetsPage';
import PerformancePage from './pages/PerformancePage';
import MonitoringPage from './pages/MonitoringPage';
import CommissionsPage from './pages/CommissionsPage';

function ProtectedRoute({ children }) {
  const { user } = useAuth();
  if (!user) return <Navigate to="/login" replace />;
  return children;
}

export default function App() {
  const { user } = useAuth();

  return (
    <Routes>
      <Route
        path="/login"
        element={user ? <Navigate to="/" replace /> : <LoginPage />}
      />
      <Route
        path="/"
        element={
          <ProtectedRoute>
            <AppLayout />
          </ProtectedRoute>
        }
      >
        <Route index element={<StudentsPage />} />
        <Route path="teachers" element={<TeachersPage />} />
        <Route path="disciplines" element={<DisciplinesPage />} />
        <Route path="sheets" element={<SheetsPage />} />
        <Route path="performance" element={<PerformancePage />} />
        <Route path="monitoring" element={<MonitoringPage />} />
        <Route path="commissions" element={<CommissionsPage />} />
      </Route>
    </Routes>
  );
}
