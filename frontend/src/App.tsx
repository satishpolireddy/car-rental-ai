import React, { useEffect } from 'react';
import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SearchPage } from './pages/SearchPage';
import { BookingPage } from './pages/BookingPage';
import { ConfirmationPage } from './pages/ConfirmationPage';
import { CustomerHomePage } from './pages/CustomerHomePage';
import { LoginPage } from './pages/LoginPage';
import { PaymentPage } from './pages/PaymentPage';
import { ProtectedRoute } from './components/ProtectedRoute';
import { useAuthStore } from './store/authStore';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: 2,
    },
  },
});

function AppRoutes() {
  const hydrate = useAuthStore((s) => s.hydrate);
  useEffect(() => { hydrate(); }, [hydrate]);

  return (
    <Routes>
      {/* Public routes */}
      <Route path="/" element={<SearchPage />} />
      <Route path="/login" element={<LoginPage />} />

      {/* Protected routes */}
      <Route path="/book/:id" element={
        <ProtectedRoute><BookingPage /></ProtectedRoute>
      } />
      <Route path="/pay/:id" element={
        <ProtectedRoute><PaymentPage /></ProtectedRoute>
      } />
      <Route path="/confirmation/:id" element={
        <ProtectedRoute><ConfirmationPage /></ProtectedRoute>
      } />
      <Route path="/my-bookings" element={
        <ProtectedRoute><CustomerHomePage /></ProtectedRoute>
      } />
    </Routes>
  );
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AppRoutes />
      </BrowserRouter>
    </QueryClientProvider>
  );
}

export default App;
