import { NavLink } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  Users, GraduationCap, BookOpen,
  ClipboardList, BarChart3, AlertTriangle, Scale
} from 'lucide-react';
import './Sidebar.css';

const navItems = [
  { to: '/', icon: <Users size={20} />, label: 'Студенты' },
  { to: '/teachers', icon: <GraduationCap size={20} />, label: 'Преподаватели' },
  { to: '/disciplines', icon: <BookOpen size={20} />, label: 'Дисциплины' },
  { to: '/sheets', icon: <ClipboardList size={20} />, label: 'Ведомость' },
  { to: '/performance', icon: <BarChart3 size={20} />, label: 'Успеваемость' },
  { to: '/monitoring', icon: <AlertTriangle size={20} />, label: 'Мониторинг' },
  { to: '/commissions', icon: <Scale size={20} />, label: 'Комиссия' },
];

export default function Sidebar() {
  return (
    <motion.aside
      className="sidebar"
      initial={{ x: -260 }}
      animate={{ x: 0 }}
      transition={{ type: 'spring', stiffness: 300, damping: 30 }}
    >
      <div className="sidebar-logo">
        <div className="sidebar-logo-icon">
          <GraduationCap size={24} />
        </div>
        <span>Кафедра</span>
      </div>

      <nav className="sidebar-nav">
        {navItems.map((item, i) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) => `sidebar-link ${isActive ? 'active' : ''}`}
          >
            <motion.div
              className="sidebar-link-inner"
              initial={{ opacity: 0, x: -20 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ delay: 0.1 + i * 0.05 }}
            >
              {item.icon}
              <span>{item.label}</span>
            </motion.div>
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-footer">
        <span className="sidebar-version">v1.0.0</span>
      </div>
    </motion.aside>
  );
}
