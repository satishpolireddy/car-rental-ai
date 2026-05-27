import { create } from 'zustand';
import { Car, SearchRequest } from '../services/api';

interface BookingStore {
  searchParams: Partial<SearchRequest>;
  selectedCar: Car | null;
  step: 'search' | 'results' | 'details' | 'confirm' | 'success';
  bookingId: number | null;

  setSearchParams: (params: Partial<SearchRequest>) => void;
  setSelectedCar: (car: Car | null) => void;
  setStep: (step: BookingStore['step']) => void;
  setBookingId: (id: number) => void;
  reset: () => void;
}

export const useBookingStore = create<BookingStore>((set) => ({
  searchParams: {},
  selectedCar: null,
  step: 'search',
  bookingId: null,

  setSearchParams: (params) => set((s) => ({ searchParams: { ...s.searchParams, ...params } })),
  setSelectedCar: (car) => set({ selectedCar: car }),
  setStep: (step) => set({ step }),
  setBookingId: (id) => set({ bookingId: id }),
  reset: () => set({ searchParams: {}, selectedCar: null, step: 'search', bookingId: null }),
}));
