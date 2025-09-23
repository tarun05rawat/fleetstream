import React, { useMemo } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { ChartDataPoint } from '../types';

interface RealtimeChartProps {
  data: ChartDataPoint[];
  title: string;
  height?: number;
  showLegend?: boolean;
  className?: string;
}

const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    const timestamp = new Date(label).toLocaleTimeString();

    return (
      <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
        <p className="text-sm font-medium text-gray-900 mb-2">{timestamp}</p>
        {payload.map((entry: any, index: number) => (
          <div key={index} className="flex items-center gap-2 text-sm">
            <div
              className="w-3 h-3 rounded-full"
              style={{ backgroundColor: entry.color }}
            />
            <span className="capitalize">{entry.dataKey.replace('_', ' ')}:</span>
            <span className="font-medium">{entry.value.toFixed(2)}</span>
            {entry.dataKey === 'temperature' && <span className="text-gray-500">°C</span>}
            {entry.dataKey === 'conveyor_speed' && <span className="text-gray-500">m/s</span>}
            {entry.dataKey === 'robot_arm_angle' && <span className="text-gray-500">°</span>}
          </div>
        ))}
      </div>
    );
  }

  return null;
};

export const RealtimeChart: React.FC<RealtimeChartProps> = ({
  data,
  title,
  height = 300,
  showLegend = true,
  className = '',
}) => {
  const chartData = useMemo(() => {
    return data.map(point => ({
      ...point,
      timestamp: new Date(point.timestamp).getTime(),
      formattedTime: new Date(point.timestamp).toLocaleTimeString(),
    })).sort((a, b) => a.timestamp - b.timestamp);
  }, [data]);

  const formatXAxisTick = (value: number) => {
    return new Date(value).toLocaleTimeString([], {
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    });
  };

  return (
    <div className={`card ${className}`}>
      <div className="card-header">
        <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
      </div>

      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={chartData} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
          <XAxis
            dataKey="timestamp"
            type="number"
            scale="time"
            domain={['dataMin', 'dataMax']}
            tickFormatter={formatXAxisTick}
            stroke="#6b7280"
            fontSize={12}
          />
          <YAxis stroke="#6b7280" fontSize={12} />
          <Tooltip content={<CustomTooltip />} />
          {showLegend && (
            <Legend
              formatter={(value) => (
                <span className="capitalize text-sm">
                  {value.replace('_', ' ')}
                </span>
              )}
            />
          )}

          <Line
            type="monotone"
            dataKey="conveyor_speed"
            stroke="#3b82f6"
            strokeWidth={2}
            dot={false}
            name="Conveyor Speed"
            connectNulls={false}
          />
          <Line
            type="monotone"
            dataKey="temperature"
            stroke="#ef4444"
            strokeWidth={2}
            dot={false}
            name="Temperature"
            connectNulls={false}
          />
          <Line
            type="monotone"
            dataKey="robot_arm_angle"
            stroke="#22c55e"
            strokeWidth={2}
            dot={false}
            name="Robot Arm Angle"
            connectNulls={false}
          />
        </LineChart>
      </ResponsiveContainer>

      {chartData.length === 0 && (
        <div className="flex items-center justify-center h-32 text-gray-500">
          No data available
        </div>
      )}
    </div>
  );
};