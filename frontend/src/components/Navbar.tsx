import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '../store/authStore';
import { logoutApi } from '../services/api';
import { auth } from '../lib/auth';

export const Navbar: React.FC = () => {
  const navigate = useNavigate();
  const { user, isLoggedIn, logout } = useAuthStore();

  const handleLogout = async () => {
    try {
      const refreshToken = auth.getRefreshToken();
      if (refreshToken) await logoutApi(refreshToken);
    } catch (_) { /* best-effort */ }
    logout();
    navigate('/');
  };

  return (
    <nav className="bg-white/90 backdrop-blur-sm border-b border-gray-200 px-4 py-3 flex items-center justify-between sticky top-0 z-50">
      <button
        onClick={() => navigate('/')}
        className="flex items-center gap-2 font-bold text-gray-900 text-lg hover:text-blue-600 transition"
      >
        🚗 <span>DriveEasy AI</span>
      </button>

      <div className="flex items-center gap-3">
        {isLoggedIn && user ? (
          <>
            <button
              onClick={() => navigate('/my-bookings')}
              className="text-sm text-gray-600 hover:text-blue-600 font-medium transition"
            >
              My Bookings
            </button>
            <div className="flex items-center gap-2 bg-gray-100 rounded-full px-3 py-1.5">
              <div className="w-6 h-6 bg-blue-600 rounded-full flex items-center justify-center text-white text-xs font-bold">
                {user.first_name?.[0]?.toUpperCase() || 'U'}
              </div>
              <span className="text-sm font-medium text-gray-800">{user.first_name}</span>
            </div>
            <button
              onClick={handleLogout}
              className="text-sm text-gray-500 hover:text-red-500 transition"
            >
              Sign out
            </button>
          </>
        ) : (
          <button
            onClick={() => navigate('/login')}
            className="bg-blue-600 hover:bg-blue-700 text-white text-sm font-semibold px-4 py-2 rounded-lg transition"
          >
            Sign in
          </button>
        )}
      </div>
    </nav>
  );
};
