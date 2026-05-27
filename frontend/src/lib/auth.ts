const ACCESS_TOKEN_KEY = 'car_rental_access_token';
const REFRESH_TOKEN_KEY = 'car_rental_refresh_token';
const USER_KEY = 'car_rental_user';

export interface AuthUser {
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  is_verified: boolean;
}

export const auth = {
  setTokens(access: string, refresh: string) {
    localStorage.setItem(ACCESS_TOKEN_KEY, access);
    localStorage.setItem(REFRESH_TOKEN_KEY, refresh);
  },
  getAccessToken(): string | null {
    return localStorage.getItem(ACCESS_TOKEN_KEY);
  },
  getRefreshToken(): string | null {
    return localStorage.getItem(REFRESH_TOKEN_KEY);
  },
  setUser(user: AuthUser) {
    localStorage.setItem(USER_KEY, JSON.stringify(user));
  },
  getUser(): AuthUser | null {
    const raw = localStorage.getItem(USER_KEY);
    return raw ? JSON.parse(raw) : null;
  },
  clear() {
    localStorage.removeItem(ACCESS_TOKEN_KEY);
    localStorage.removeItem(REFRESH_TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
  },
  isLoggedIn(): boolean {
    return !!localStorage.getItem(ACCESS_TOKEN_KEY);
  },
};
