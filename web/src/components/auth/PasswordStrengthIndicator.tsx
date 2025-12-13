import type { FC } from 'react';

interface Props {
  password: string;
}

const scorePassword = (pwd: string): number => {
  let score = 0;
  if (!pwd) return score;
  if (pwd.length >= 8) score++;
  if (pwd.length >= 12) score++;
  if (/[A-Z]/.test(pwd)) score++;
  if (/[0-9]/.test(pwd)) score++;
  if (/[^A-Za-z0-9]/.test(pwd)) score++;
  return Math.min(score, 4);
};

export const PasswordStrengthIndicator: FC<Props> = ({ password }) => {
  const score = scorePassword(password);
  const labels = ['Very weak', 'Weak', 'Medium', 'Strong', 'Very strong'];
  const colors = ['bg-red-500', 'bg-orange-500', 'bg-yellow-500', 'bg-green-500', 'bg-emerald-600'];

  if (!password) return null;

  return (
    <div className="mt-2">
      <div className="flex gap-1 mb-1">
        {Array.from({ length: 5 }).map((_, i) => (
          <div
            key={i}
            className={`h-1 flex-1 rounded-full ${i <= score ? colors[score] : 'bg-gray-200'}`}
          />
        ))}
      </div>
      <p className="text-xs text-gray-600">Password strength: {labels[score]}</p>
    </div>
  );
};

export default PasswordStrengthIndicator;
