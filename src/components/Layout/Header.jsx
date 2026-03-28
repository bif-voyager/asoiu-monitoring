import { useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import { LogOut, User } from 'lucide-react';
import './Header.css';

const titles = {
  '/': 'Студенты',
  '/teachers': 'Преподаватели',
  '/disciplines': 'Дисциплины',
  '/sheets': 'Ведомость',
  '/performance': 'Успеваемость',
  '/monitoring': 'Мониторинг',
  '/commissions': 'Комиссия',
};

export default function Header() {
  const { user, logout } = useAuth();
  const location = useLocation();
  const title = titles[location.pathname] || 'Мониторинг успеваемости';

  return (
    <header className="app-header">
      <h1 className="header-title">{title}</h1>
      <div className="header-right">
        <div className="header-user">
          <div className="header-user-avatar">
            <User size={16} />
          </div>
          <div className="header-user-info">
            <span className="header-user-name">{user?.name || 'Гость'}</span>
            <span className="header-user-role">{user?.role || ''}</span>
          </div>
        </div>
        <button className="btn-icon" onClick={logout} title="Выйти">
          <LogOut size={20} />
        </button>
      </div>
    </header>
  );
}
