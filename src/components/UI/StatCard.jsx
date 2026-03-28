import { motion } from 'framer-motion';
import './StatCard.css';

export default function StatCard({ icon, label, value, color = 'accent', delay = 0 }) {
  return (
    <motion.div
      className={`stat-card stat-card-${color}`}
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay, duration: 0.4 }}
    >
      <div className="stat-icon">{icon}</div>
      <div className="stat-info">
        <span className="stat-value">{value}</span>
        <span className="stat-label">{label}</span>
      </div>
    </motion.div>
  );
}
