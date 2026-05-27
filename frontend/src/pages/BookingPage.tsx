import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import { getCar, createBooking } from '../services/api';
import { useBookingStore } from '../store/bookingStore';
import { useAuthStore } from '../store/authStore';
import { differenceInCalendarDays, format } from 'date-fns';
import { Navbar } from '../components/Navbar';

export const BookingPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { searchParams, setBookingId } = useBookingStore();
  const user = useAuthStore((s) => s.user);

  const { data: car, isLoading } = useQuery({
    queryKey: ['car', id],
    queryFn: () => getCar(Number(id)),
  });

  const [notes, setNotes] = useState('');
  const [error, setError] = useState('');

  const days =
    searchParams.pickup_date && searchParams.return_date
      ? differenceInCalendarDays(
          new Date(searchParams.return_date),
          new Date(searchParams.pickup_date)
        )
      : 1;

  const totalAmount = car ? car.daily_rate * days : 0;

  // After booking is created → go to payment page
  const mutation = useMutation({
    mutationFn: createBooking,
    onSuccess: (booking) => {
      setBookingId(booking.id);
      navigate(`/pay/${booking.id}`);
    },
    onError: (err: any) => {
      setError(err?.response?.data?.error || 'Booking failed. Please try again.');
    },
  });

  const handleConfirm = () => {
    if (!car || !searchParams.pickup_date || !searchParams.return_date || !user) return;
    setError('');
    mutation.mutate({
      customer_id: user.id,
      car_id: car.id,
      pickup_date: new Date(searchParams.pickup_date).toISOString(),
      return_date: new Date(searchParams.return_date).toISOString(),
      pickup_location: searchParams.location || '',
      return_location: searchParams.location || '',
      notes,
    });
  };

  if (isLoading)
    return (
      <>
        <Navbar />
        <div className="p-8 text-center text-gray-500">Loading...</div>
      </>
    );
  if (!car)
    return (
      <>
        <Navbar />
        <div className="p-8 text-center text-red-500">Car not found</div>
      </>
    );

  return (
    <>
      <Navbar />
      <div className="min-h-screen bg-gray-50 py-10 px-4">
        <div className="max-w-2xl mx-auto">
          <button onClick={() => navigate(-1)} className="text-blue-600 text-sm mb-6 hover:underline">
            ← Back to results
          </button>
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Confirm Your Booking</h2>

          {/* Car Summary */}
          <div className="bg-white rounded-2xl border border-gray-200 p-6 mb-6 shadow-sm">
            <h3 className="font-bold text-lg mb-1">
              {car.year} {car.make} {car.model}
            </h3>
            <p className="text-gray-500 text-sm capitalize mb-4">
              {car.category} · {car.location}
            </p>
            <div className="grid grid-cols-3 gap-3 text-sm text-gray-600">
              <div>👥 {car.seats} seats</div>
              <div>⚙️ {car.transmission}</div>
              <div>⛽ {car.fuel_type}</div>
            </div>
          </div>

          {/* Booking Details */}
          <div className="bg-white rounded-2xl border border-gray-200 p-6 mb-6 shadow-sm">
            <h4 className="font-semibold mb-4">Trip Details</h4>
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <p className="text-gray-500 text-xs">PICK-UP</p>
                <p className="font-medium">
                  {searchParams.pickup_date
                    ? format(new Date(searchParams.pickup_date), 'EEE, MMM d yyyy')
                    : '—'}
                </p>
                <p className="text-gray-600">{searchParams.location}</p>
              </div>
              <div>
                <p className="text-gray-500 text-xs">RETURN</p>
                <p className="font-medium">
                  {searchParams.return_date
                    ? format(new Date(searchParams.return_date), 'EEE, MMM d yyyy')
                    : '—'}
                </p>
                <p className="text-gray-600">{searchParams.location}</p>
              </div>
            </div>
            <div className="border-t mt-4 pt-4">
              <div className="flex justify-between text-sm mb-1">
                <span className="text-gray-500">
                  ${car.daily_rate}/day × {days} days
                </span>
                <span>${totalAmount.toFixed(2)}</span>
              </div>
              <div className="flex justify-between font-bold text-lg mt-2">
                <span>Total</span>
                <span className="text-blue-600">${totalAmount.toFixed(2)}</span>
              </div>
            </div>
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-gray-700 mb-1">
              Special requests (optional)
            </label>
            <textarea
              value={notes}
              onChange={(e) => setNotes(e.target.value)}
              rows={3}
              placeholder="Child seat, GPS, early pick-up..."
              className="w-full border border-gray-300 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
            />
          </div>

          {error && <p className="text-red-600 text-sm mb-4">{error}</p>}

          <button
            onClick={handleConfirm}
            disabled={mutation.isPending}
            className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white font-bold py-3 px-6 rounded-xl transition-colors text-lg"
          >
            {mutation.isPending
              ? 'Creating booking…'
              : `Continue to Payment · $${totalAmount.toFixed(2)}`}
          </button>
          <p className="text-center text-xs text-gray-400 mt-2">
            You'll be taken to a secure payment page
          </p>
        </div>
      </div>
    </>
  );
};
