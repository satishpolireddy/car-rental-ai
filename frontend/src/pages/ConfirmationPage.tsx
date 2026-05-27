import React from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getBooking } from '../services/api';
import { format } from 'date-fns';

export const ConfirmationPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: booking, isLoading } = useQuery({
    queryKey: ['booking', id],
    queryFn: () => getBooking(Number(id)),
  });

  if (isLoading) return <div className="p-8 text-center text-gray-500">Loading booking...</div>;
  if (!booking) return <div className="p-8 text-center text-red-500">Booking not found</div>;

  return (
    <div className="min-h-screen bg-gray-50 py-10 px-4">
      <div className="max-w-lg mx-auto text-center">
        <div className="text-6xl mb-4">🎉</div>
        <h2 className="text-3xl font-bold text-gray-900 mb-2">Booking Confirmed!</h2>
        <p className="text-gray-500 mb-8">Booking #{booking.id} · <span className="capitalize text-green-600 font-medium">{booking.status}</span></p>

        <div className="bg-white rounded-2xl border border-gray-200 p-6 text-left shadow-sm mb-6">
          {booking.car && (
            <div className="mb-4 pb-4 border-b">
              <p className="font-bold text-lg">{booking.car.year} {booking.car.make} {booking.car.model}</p>
              <p className="text-gray-500 text-sm capitalize">{booking.car.category}</p>
            </div>
          )}
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-gray-500 text-xs">PICK-UP</p>
              <p className="font-medium">{format(new Date(booking.pickup_date), 'EEE, MMM d yyyy')}</p>
              <p className="text-gray-600 text-xs">{booking.pickup_location}</p>
            </div>
            <div>
              <p className="text-gray-500 text-xs">RETURN</p>
              <p className="font-medium">{format(new Date(booking.return_date), 'EEE, MMM d yyyy')}</p>
              <p className="text-gray-600 text-xs">{booking.return_location}</p>
            </div>
          </div>
          <div className="border-t mt-4 pt-4 flex justify-between font-bold">
            <span>{booking.total_days} days</span>
            <span className="text-blue-600">${booking.total_amount.toFixed(2)}</span>
          </div>
        </div>

        <div className="flex gap-3">
          <button
            onClick={() => navigate('/')}
            className="flex-1 bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-6 rounded-xl transition-colors"
          >
            Book Another Car
          </button>
          <button
            onClick={() => navigate('/my-bookings')}
            className="flex-1 border border-gray-300 hover:bg-gray-50 text-gray-700 font-bold py-3 px-6 rounded-xl transition-colors"
          >
            My Bookings
          </button>
        </div>
      </div>
    </div>
  );
};
