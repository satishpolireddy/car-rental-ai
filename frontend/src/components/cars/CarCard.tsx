import React from 'react';
import { Car } from '../../services/api';

interface CarCardProps {
  car: Car;
  onSelect: (car: Car) => void;
  highlighted?: boolean;
}

const categoryIcons: Record<string, string> = {
  economy: '🚗',
  standard: '🚙',
  luxury: '🏎️',
  suv: '🚐',
  van: '🚌',
};

const fuelIcons: Record<string, string> = {
  electric: '⚡',
  hybrid: '🔋',
  petrol: '⛽',
  diesel: '⛽',
};

export const CarCard: React.FC<CarCardProps> = ({ car, onSelect, highlighted }) => {
  return (
    <div
      className={`rounded-2xl border bg-white shadow-sm hover:shadow-md transition-all cursor-pointer p-5
        ${highlighted ? 'border-blue-500 ring-2 ring-blue-200' : 'border-gray-200'}`}
      onClick={() => onSelect(car)}
    >
      {highlighted && (
        <div className="mb-2 inline-block bg-blue-100 text-blue-700 text-xs font-semibold px-2 py-1 rounded-full">
          ✨ AI Recommended
        </div>
      )}
      <div className="flex items-start justify-between mb-3">
        <div>
          <h3 className="text-lg font-bold text-gray-900">
            {categoryIcons[car.category] || '🚗'} {car.year} {car.make} {car.model}
          </h3>
          <p className="text-sm text-gray-500 capitalize">{car.category} · {car.location}</p>
        </div>
        <div className="text-right">
          <p className="text-2xl font-bold text-blue-600">${car.daily_rate}</p>
          <p className="text-xs text-gray-400">per day</p>
        </div>
      </div>

      <div className="grid grid-cols-3 gap-2 text-sm text-gray-600 mb-4">
        <div className="flex items-center gap-1">
          <span>👥</span> {car.seats} seats
        </div>
        <div className="flex items-center gap-1 capitalize">
          <span>⚙️</span> {car.transmission}
        </div>
        <div className="flex items-center gap-1">
          <span>{fuelIcons[car.fuel_type] || '⛽'}</span> {car.fuel_type}
        </div>
      </div>

      <button
        className="w-full bg-blue-600 hover:bg-blue-700 text-white font-semibold py-2 px-4 rounded-xl transition-colors"
        onClick={(e) => { e.stopPropagation(); onSelect(car); }}
      >
        Select This Car
      </button>
    </div>
  );
};
