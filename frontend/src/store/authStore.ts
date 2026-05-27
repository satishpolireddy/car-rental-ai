import { create } from 'zustand';
import { auth, AuthUser } from '../lib/auth';

interface AuthState {
  user: AuthUser | null;
  isLoggedIn: boolean;
  setAuth: (user: AuthUser, accessToken: string, refreshToken: string) => void;
  logout: () => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  isLoggedIn: false,

  setAuth: (user, accessToken, refreshToken) => {
    auth.setTokens(accessToken, refreshToken);
    auth.setUser(user);
    set({ user, isLoggedIn: true });
  },

  logout: () => {
    auth.clear();
    set({ user: null, isLoggedIn: false });
  },

  hydrate: () => {
    const user = auth.getUser();
    const token = auth.getAccessToken();
    if (user && token) {
      set({ user, isLoggedIn: true });
    }
  },
}));
