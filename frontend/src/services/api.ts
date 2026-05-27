import axios from 'axios';
import { auth } from '../lib/auth';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

// Attach JWT to every request
api.interceptors.request.use((config) => {
  const token = auth.getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Auto-refresh on 401
let isRefreshing = false;
let pendingQueue: Array<{ resolve: (t: string) => void; reject: (e: unknown) => void }> = [];

const processQueue = (error: unknown, token: string | null = null) => {
  pendingQueue.forEach((p) => (error ? p.reject(error) : p.resolve(token!)));
  pendingQueue = [];
};

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config;
    if (error.response?.status === 401 && !original._retry) {
      const refreshToken = auth.getRefreshToken();
      if (!refreshToken) {
        auth.clear();
        window.location.href = '/login';
        return Promise.reject(error);
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingQueue.push({ resolve, reject });
        }).then((token) => {
          original.headers.Authorization = `Bearer ${token}`;
          return api(original);
        });
      }

      original._retry = true;
      isRefreshing = true;

      try {
        const { data } = await axios.post(
          `${api.defaults.baseURL}/auth/refresh`,
          { refresh_token: refreshToken }
        );
        auth.setTokens(data.access_token, data.refresh_token);
        processQueue(null, data.access_token);
        original.headers.Authorization = `Bearer ${data.access_token}`;
        return api(original);
      } catch (err) {
        processQueue(err, null);
        auth.clear();
        window.location.href = '/login';
        return Promise.reject(err);
      } finally {
        isRefreshing = false;
      }
    }
    return Promise.reject(error);
  }
);

// ─── Types ────────────────────────────────────────────────────────────────────

export interface Car {
  id: number;
  make: string;
  model: string;
  year: number;
  category: string;
  daily_rate: number;
  available: boolean;
  location: string;
  seats: number;
  transmission: string;
  fuel_type: string;
  features: string;
  image_url: string;
}

export interface SearchRequest {
  location: string;
  pickup_date: string;
  return_date: string;
  category?: string;
  max_daily_rate?: number;
  min_seats?: number;
}

export interface BookingRequest {
  customer_id: number;
  car_id: number;
  pickup_date: string;
  return_date: string;
  pickup_location: string;
  return_location: string;
  notes?: string;
}

export interface Booking {
  id: number;
  customer_id: number;
  car_id: number;
  pickup_date: string;
  return_date: string;
  pickup_location: string;
  return_location: string;
  total_days: number;
  total_amount: number;
  status: string;
  car?: Car;
}

export interface AIResponse {
  message: string;
  recommended_car_ids: number[];
  latency_ms: number;
  tokens_used: number;
}

export interface Customer {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  license_number: string;
  created_at: string;
}

export interface CustomerStats {
  active_bookings: number;
  completed_bookings: number;
  cancelled_bookings: number;
  total_spent: number;
}

export interface CustomerDashboard {
  customer: Customer;
  stats: CustomerStats;
  current_bookings: Booking[];
  past_bookings: Booking[];
}

export interface PaymentIntent {
  id: number;
  booking_id: number;
  amount_cents: number;
  amount_display: string;
  currency: string;
  status: string;
  client_secret: string;
  payment_intent_id: string;
}

// ─── API calls ────────────────────────────────────────────────────────────────

export const searchCars = async (req: SearchRequest) => {
  const { data } = await api.post('/api/v1/cars/search', req);
  return data as { cars: Car[]; total: number };
};

export const getCar = async (id: number) => {
  const { data } = await api.get(`/api/v1/cars/${id}`);
  return data as Car;
};

export const getLocations = async () => {
  const { data } = await api.get('/api/v1/locations');
  return data.locations as string[];
};

export const createBooking = async (req: BookingRequest) => {
  const { data } = await api.post('/api/v1/bookings', req);
  return data as Booking;
};

export const getBooking = async (id: number) => {
  const { data } = await api.get(`/api/v1/bookings/${id}`);
  return data as Booking;
};

export const cancelBooking = async (id: number) => {
  await api.delete(`/api/v1/bookings/${id}`);
};

export const getCustomerDashboard = async (customerId: number) => {
  const { data } = await api.get(`/api/v1/customers/${customerId}/dashboard`);
  return data as CustomerDashboard;
};

export const getAIRecommendation = async (query: string, customerId?: number) => {
  const { data } = await api.post('/api/v1/ai/recommend', { query, customer_id: customerId });
  return data as AIResponse;
};

// Auth
export const sendOTP = async (email: string, firstName: string, lastName: string) => {
  const { data } = await axios.post(`${api.defaults.baseURL}/auth/send-otp`, {
    email,
    first_name: firstName,
    last_name: lastName,
  });
  return data as { message: string; is_new_user: boolean };
};

export const verifyOTP = async (email: string, otp: string) => {
  const { data } = await axios.post(`${api.defaults.baseURL}/auth/verify-otp`, { email, otp });
  return data as {
    access_token: string;
    refresh_token: string;
    token_type: string;
    user: import('../lib/auth').AuthUser;
  };
};

export const logoutApi = async (refreshToken: string) => {
  await api.post('/auth/logout', { refresh_token: refreshToken });
};

// Payments
export const createPaymentIntent = async (
  bookingId: number,
  customerId: number,
  amountCents: number
) => {
  const { data } = await api.post('/payments', {
    booking_id: bookingId,
    customer_id: customerId,
    amount_cents: amountCents,
  });
  return data as PaymentIntent;
};

export const getPaymentByBooking = async (bookingId: number) => {
  const { data } = await api.get(`/payments/booking/${bookingId}`);
  return data as PaymentIntent;
};

export default api;
