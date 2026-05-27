import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { getCustomerDashboard, cancelBooking, Booking } from '../services/api';
import { format, differenceInDays, isPast, isFuture } from 'date-fns';

// Hardcoded to customer 1 for demo — replace with auth context in production
const CUSTOMER_ID = 1;

// ─── Status badge ──────────────────────────────────────────────────────────────
const StatusBadge: React.FC<{ status: string }> = ({ status }) => {
  const styles: Record<string, string> = {
    confirmed:  'bg-blue-100 text-blue-700',
    active:     'bg-green-100 text-green-700',
    completed:  'bg-gray-100 text-gray-600',
    cancelled:  'bg-red-100 text-red-500',
  };
  const icons: Record<string, string> = {
    confirmed: '📋', active: '🚗', completed: '✅', cancelled: '❌',
  };
  return (
    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-semibold capitalize ${styles[status] || 'bg-gray-100 text-gray-500'}`}>
      {icons[status]} {status}
    </span>
  );
};

// ─── Booking card ──────────────────────────────────────────────────────────────
const BookingCard: React.FC<{
  booking: Booking;
  onCancel?: (id: number) => void;
  cancelling?: boolean;
}> = ({ booking, onCancel, cancelling }) => {
  const pickupDate = new Date(booking.pickup_date);
  const returnDate = new Date(booking.return_date);
  const daysUntilPickup = differenceInDays(pickupDate, new Date());
  const isUpcoming = isFuture(pickupDate);
  const canCancel = booking.status === 'confirmed' || booking.status === 'active';

  return (
    <div className="bg-white rounded-2xl border border-gray-200 shadow-sm hover:shadow-md transition-shadow p-5">
      {/* Header row */}
      <div className="flex items-start justify-between mb-3">
        <div>
          <p className="text-xs text-gray-400 font-medium mb-0.5">BOOKING #{booking.id}</p>
          {booking.car ? (
            <h3 className="font-bold text-gray-900 text-base">
              {booking.car.year} {booking.car.make} {booking.car.model}
            </h3>
          ) : (
            <h3 className="font-bold text-gray-900">Car #{booking.car_id}</h3>
          )}
          {booking.car && (
            <p className="text-xs text-gray-500 capitalize mt-0.5">
              {booking.car.category} · {booking.car.transmission} · {booking.car.fuel_type}
            </p>
          )}
        </div>
        <StatusBadge status={booking.status} />
      </div>

      {/* Dates */}
      <div className="grid grid-cols-2 gap-3 mb-3 bg-gray-50 rounded-xl p-3 text-sm">
        <div>
          <p className="text-xs text-gray-400 font-medium">PICK-UP</p>
          <p className="font-semibold text-gray-800">{format(pickupDate, 'EEE, MMM d yyyy')}</p>
          <p className="text-xs text-gray-500 mt-0.5">📍 {booking.pickup_location}</p>
        </div>
        <div>
          <p className="text-xs text-gray-400 font-medium">RETURN</p>
          <p className="font-semibold text-gray-800">{format(returnDate, 'EEE, MMM d yyyy')}</p>
          <p className="text-xs text-gray-500 mt-0.5">📍 {booking.return_location}</p>
        </div>
      </div>

      {/* Countdown for upcoming bookings */}
      {isUpcoming && booking.status === 'confirmed' && daysUntilPickup >= 0 && (
        <div className="mb-3 bg-blue-50 text-blue-700 rounded-xl px-3 py-2 text-xs font-medium">
          ⏰ {daysUntilPickup === 0 ? 'Pick-up is today!' : `Pick-up in ${daysUntilPickup} day${daysUntilPickup !== 1 ? 's' : ''}`}
        </div>
      )}

      {/* Footer: duration + cost + action */}
      <div className="flex items-center justify-between pt-3 border-t border-gray-100">
        <div className="text-sm text-gray-600">
          <span className="font-medium">{booking.total_days} day{booking.total_days !== 1 ? 's' : ''}</span>
          <span className="text-gray-400 mx-1">·</span>
          <span className="font-bold text-blue-600">${booking.total_amount.toFixed(2)}</span>
        </div>
        {canCancel && onCancel && (
          <button
            onClick={() => onCancel(booking.id)}
            disabled={cancelling}
            className="text-xs text-red-500 hover:text-red-700 border border-red-200 hover:border-red-400 px-3 py-1 rounded-lg transition-colors disabled:opacity-50"
          >
            {cancelling ? 'Cancelling...' : 'Cancel'}
          </button>
        )}
      </div>
    </div>
  );
};

// ─── Stat card ─────────────────────────────────────────────────────────────────
const StatCard: React.FC<{ label: string; value: string | number; icon: string; color: string }> = ({
  label, value, icon, color,
}) => (
  <div className={`rounded-2xl p-4 ${color}`}>
    <div className="text-2xl mb-1">{icon}</div>
    <div className="text-2xl font-bold text-gray-900">{value}</div>
    <div className="text-xs text-gray-500 font-medium mt-0.5">{label}</div>
  </div>
);

// ─── Main page ─────────────────────────────────────────────────────────────────
export const CustomerHomePage: React.FC = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [cancellingId, setCancellingId] = useState<number | null>(null);
  const [tab, setTab] = useState<'current' | 'past'>('current');

  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard', CUSTOMER_ID],
    queryFn: () => getCustomerDashboard(CUSTOMER_ID),
  });

  const cancelMutation = useMutation({
    mutationFn: cancelBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['dashboard', CUSTOMER_ID] });
      setCancellingId(null);
    },
    onError: () => setCancellingId(null),
  });

  const handleCancel = (id: number) => {
    if (!window.confirm('Are you sure you want to cancel this booking?')) return;
    setCancellingId(id);
    cancelMutation.mutate(id);
  };

  if (isLoading) return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center text-gray-500">
        <div className="text-4xl mb-3 animate-pulse">🚗</div>
        <p>Loading your dashboard...</p>
      </div>
    </div>
  );

  if (error || !data) return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center">
      <div className="text-center text-red-500">
        <div className="text-4xl mb-3">⚠️</div>
        <p>Failed to load dashboard. Please try again.</p>
      </div>
    </div>
  );

  const { customer, stats, current_bookings, past_bookings } = data;
  const activeList = tab === 'current' ? current_bookings : past_bookings;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Top nav */}
      <div className="bg-white border-b border-gray-200 px-4 py-3">
        <div className="max-w-3xl mx-auto flex items-center justify-between">
          <button onClick={() => navigate('/')} className="text-blue-600 text-sm font-medium hover:underline">
            ← Back to Search
          </button>
          <span className="text-lg font-bold text-gray-800">🚗 DriveAI</span>
          <button
            onClick={() => navigate('/')}
            className="bg-blue-600 text-white text-sm font-semibold px-4 py-1.5 rounded-xl hover:bg-blue-700 transition-colors"
          >
            + New Booking
          </button>
        </div>
      </div>

      <div className="max-w-3xl mx-auto px-4 py-8">
        {/* Profile header */}
        <div className="bg-gradient-to-r from-blue-600 to-indigo-600 rounded-2xl p-6 text-white mb-6">
          <div className="flex items-center gap-4">
            <div className="w-14 h-14 rounded-full bg-white/20 flex items-center justify-center text-2xl font-bold">
              {customer.first_name[0]}{customer.last_name[0]}
            </div>
            <div>
              <h1 className="text-xl font-bold">
                Welcome back, {customer.first_name}!
              </h1>
              <p className="text-blue-200 text-sm mt-0.5">{customer.email}</p>
              {customer.phone && (
                <p className="text-blue-200 text-xs mt-0.5">📞 {customer.phone}</p>
              )}
            </div>
          </div>
        </div>

        {/* Stats row */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-3 mb-6">
          <StatCard
            icon="🚗" label="Active Bookings"
            value={stats.active_bookings} color="bg-blue-50"
          />
          <StatCard
            icon="✅" label="Completed Trips"
            value={stats.completed_bookings} color="bg-green-50"
          />
          <StatCard
            icon="❌" label="Cancelled"
            value={stats.cancelled_bookings} color="bg-red-50"
          />
          <StatCard
            icon="💰" label="Total Spent"
            value={`$${stats.total_spent.toFixed(2)}`} color="bg-yellow-50"
          />
        </div>

        {/* Tab switcher */}
        <div className="flex gap-2 mb-4 bg-gray-100 rounded-xl p-1 w-fit">
          <button
            onClick={() => setTab('current')}
            className={`px-5 py-2 rounded-lg text-sm font-semibold transition-all ${
              tab === 'current'
                ? 'bg-white text-blue-600 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            Current Bookings
            {current_bookings.length > 0 && (
              <span className="ml-2 bg-blue-100 text-blue-600 text-xs px-1.5 py-0.5 rounded-full">
                {current_bookings.length}
              </span>
            )}
          </button>
          <button
            onClick={() => setTab('past')}
            className={`px-5 py-2 rounded-lg text-sm font-semibold transition-all ${
              tab === 'past'
                ? 'bg-white text-blue-600 shadow-sm'
                : 'text-gray-500 hover:text-gray-700'
            }`}
          >
            Past Bookings
            {past_bookings.length > 0 && (
              <span className="ml-2 bg-gray-200 text-gray-600 text-xs px-1.5 py-0.5 rounded-full">
                {past_bookings.length}
              </span>
            )}
          </button>
        </div>

        {/* Booking list */}
        {activeList.length === 0 ? (
          <div className="text-center py-14 text-gray-400">
            <div className="text-5xl mb-3">{tab === 'current' ? '📋' : '🗂️'}</div>
            <p className="text-base font-medium">
              {tab === 'current' ? 'No active bookings' : 'No past bookings yet'}
            </p>
            {tab === 'current' && (
              <button
                onClick={() => navigate('/')}
                className="mt-4 bg-blue-600 text-white font-semibold px-6 py-2 rounded-xl hover:bg-blue-700 transition-colors text-sm"
              >
                Book a Car Now
              </button>
            )}
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {activeList.map((booking) => (
              <BookingCard
                key={booking.id}
                booking={booking}
                onCancel={tab === 'current' ? handleCancel : undefined}
                cancelling={cancellingId === booking.id}
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};
