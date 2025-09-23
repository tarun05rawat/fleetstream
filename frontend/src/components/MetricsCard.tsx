import React from 'react';
import { TrendingUp, TrendingDown, Minus } from 'lucide-react';

interface MetricsCardProps {
  title: string;
  value: string | number;
  unit?: string;
  previousValue?: number;
  trend?: 'up' | 'down' | 'stable';
  status?: 'normal' | 'warning' | 'critical';
  icon?: React.ReactNode;
  className?: string;
}

export const MetricsCard: React.FC<MetricsCardProps> = ({
  title,
  value,
  unit = '',
  previousValue,
  trend,
  status = 'normal',
  icon,
  className = '',
}) => {
  const getStatusColor = () => {
    switch (status) {
      case 'warning':
        return 'border-l-warning-500';
      case 'critical':
        return 'border-l-danger-500';
      default:
        return 'border-l-primary-500';
    }
  };

  const getTrendIcon = () => {
    switch (trend) {
      case 'up':
        return <TrendingUp size={16} className="text-success-600" />;
      case 'down':
        return <TrendingDown size={16} className="text-danger-600" />;
      case 'stable':
        return <Minus size={16} className="text-gray-400" />;
      default:
        return null;
    }
  };

  const getTrendColor = () => {
    switch (trend) {
      case 'up':
        return 'text-success-600';
      case 'down':
        return 'text-danger-600';
      default:
        return 'text-gray-500';
    }
  };

  const calculatePercentageChange = () => {
    if (previousValue && typeof value === 'number') {
      const change = ((value - previousValue) / previousValue) * 100;
      return Math.abs(change).toFixed(1);
    }
    return null;
  };

  const percentageChange = calculatePercentageChange();

  return (
    <div className={`metric-card border-l-4 ${getStatusColor()} ${className}`}>
      <div className="flex items-center justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2">
            {icon && <div className="text-gray-400">{icon}</div>}
            <h3 className="text-sm font-medium text-gray-600">{title}</h3>
          </div>
          <div className="mt-1 flex items-baseline gap-1">
            <span className="text-2xl font-bold text-gray-900">
              {typeof value === 'number' ? value.toFixed(2) : value}
            </span>
            {unit && <span className="text-sm text-gray-500">{unit}</span>}
          </div>
        </div>

        {trend && (
          <div className="flex flex-col items-end">
            {getTrendIcon()}
            {percentageChange && (
              <span className={`text-xs font-medium ${getTrendColor()}`}>
                {percentageChange}%
              </span>
            )}
          </div>
        )}
      </div>

      {status !== 'normal' && (
        <div className="mt-2">
          <span className={`text-xs font-medium ${
            status === 'warning' ? 'text-warning-700' : 'text-danger-700'
          }`}>
            {status === 'warning' ? 'Warning threshold exceeded' : 'Critical threshold exceeded'}
          </span>
        </div>
      )}
    </div>
  );
};