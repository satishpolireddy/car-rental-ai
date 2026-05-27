import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { loadStripe } from '@stripe/stripe-js';
import {
  Elements,
  PaymentElement,
  useStripe,
  useElements,
} from '@stripe/react-stripe-js';
import { useQuery } from '@tanstack/react-query';
import { getBooking, createPaymentIntent, PaymentIntent } from '../services/api';
import { useAuthStore } from '../store/authStore';

const stripePromise = loadStripe(import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY || '');

// ─── Inner checkout form (must be inside <Elements>) ─────────────────────────
const CheckoutForm: React.FC<{ payment: PaymentIntent; bookingId: number }> = ({
  payment,
  bookingId,
}) => {
  const stripe = useStripe();
  const elements = useElements();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!stripe || !elements) return;
    setLoading(true);
    setError('');

    const { error: stripeError } = await stripe.confirmPayment({
      elements,
      confirmParams: {
        return_url: `${window.location.origin}/confirmation/${bookingId}`,
      },
    });

    // If we reach here, confirmPayment failed immediately (not redirect-based)
    if (stripeError) {
      setError(stripeError.message || 'Payment failed. Please try again.');
    }
    setLoading(false);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="bg-white rounded-xl border border-gray-200 p-4">
        <PaymentElement />
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg text-sm">
          {error}
        </div>
      )}

      <button
        type="submit"
        disabled={!stripe || loading}
        className="w-full bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white font-semibold py-4 rounded-xl text-lg transition"
      >
        {loading ? 'Processing…' : `Pay ${payment.amount_display}`}
      </button>

      <p className="text-center text-xs text-gray-400">
        🔒 Secured by Stripe · Your card details are never stored on our servers
      </p>
    </form>
  );
};

// ─── Page wrapper ─────────────────────────────────────────────────────────────
export const PaymentPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const bookingId = Number(id);
  const user = useAuthStore((s) => s.user);
  const [paymentIntent, setPaymentIntent] = useState<PaymentIntent | null>(null);
  const [intentLoading, setIntentLoading] = useState(true);
  const [intentError, setIntentError] = useState('');

  const { data: booking, isLoading: bookingLoading } = useQuery({
    queryKey: ['booking', bookingId],
    queryFn: () => getBooking(bookingId),
    enabled: !!bookingId,
  });

  useEffect(() => {
    if (!booking || !user) return;
    const amountCents = Math.round(booking.total_amount * 100);
    createPaymentIntent(booking.id, user.id, amountCents)
      .then(setPaymentIntent)
      .catch((err) =>
        setIntentError(err.response?.data?.error || 'Could not initialize payment')
      )
      .finally(() => setIntentLoading(false));
  }, [booking, user]);

  if (bookingLoading || intentLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="text-5xl mb-4 animate-pulse">💳</div>
          <p className="text-gray-500">Setting up secure payment…</p>
        </div>
      </div>
    );
  }

  if (intentError || !booking || !paymentIntent) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-xl p-8 text-center max-w-sm">
          <div className="text-4xl mb-4">⚠️</div>
          <p className="text-red-600 font-medium">{intentError || 'Booking not found'}</p>
        </div>
      </div>
    );
  }

  const options = {
    clientSecret: paymentIntent.client_secret,
    appearance: { theme: 'stripe' as const },
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-blue-50 py-12 px-4">
      <div className="max-w-lg mx-auto">
        {/* Header */}
        <div className="text-center mb-8">
          <div className="text-4xl mb-2">🚗</div>
          <h1 className="text-2xl font-bold text-gray-900">Complete your booking</h1>
          <p className="text-gray-500 text-sm mt-1">Booking #{booking.id}</p>
        </div>

        {/* Summary card */}
        <div className="bg-white rounded-xl border border-gray-200 p-5 mb-6">
          <h2 className="font-semibold text-gray-800 mb-3">Booking summary</h2>
          <div className="space-y-2 text-sm text-gray-600">
            <div className="flex justify-between">
              <span>Vehicle</span>
              <span className="font-medium text-gray-800">
                {booking.car
                  ? `${booking.car.year} ${booking.car.make} ${booking.car.model}`
                  : `Car #${booking.car_id}`}
              </span>
            </div>
            <div className="flex justify-between">
              <span>Pick-up</span>
              <span>{new Date(booking.pickup_date).toLocaleDateString()}</span>
            </div>
            <div className="flex justify-between">
              <span>Return</span>
              <span>{new Date(booking.return_date).toLocaleDateString()}</span>
            </div>
            <div className="flex justify-between">
              <span>Duration</span>
              <span>{booking.total_days} day{booking.total_days !== 1 ? 's' : ''}</span>
            </div>
            <div className="border-t pt-2 mt-2 flex justify-between font-semibold text-gray-900">
              <span>Total</span>
              <span className="text-blue-600 text-lg">{paymentIntent.amount_display}</span>
            </div>
          </div>
        </div>

        {/* Stripe Elements form */}
        <Elements stripe={stripePromise} options={options}>
          <CheckoutForm payment={paymentIntent} bookingId={bookingId} />
        </Elements>
      </div>
    </div>
  );
};
