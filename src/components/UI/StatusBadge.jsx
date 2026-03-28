import './StatusBadge.css';

const statusConfig = {
  'Активен': { color: 'success' },
  'Отчислен': { color: 'danger' },
  'Академический отпуск': { color: 'warning' },
  'Назначена': { color: 'info' },
  'Проведена': { color: 'warning' },
  'Закрыта': { color: 'success' },
  'Сдано': { color: 'success' },
  'Не сдано': { color: 'danger' },
  'Отлично': { color: 'success' },
  'Хорошо': { color: 'accent' },
  'Удовлетворительно': { color: 'warning' },
  'Неудовлетворительно': { color: 'danger' },
  'Зачтено': { color: 'success' },
  'Не зачтено': { color: 'danger' },
  'Экзамен': { color: 'info' },
  'Зачёт': { color: 'accent' },
};

export default function StatusBadge({ status }) {
  const config = statusConfig[status] || { color: 'info' };
  return (
    <span className={`status-badge status-${config.color}`}>
      {status}
    </span>
  );
}
