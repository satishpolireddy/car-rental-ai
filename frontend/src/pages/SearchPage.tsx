import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { searchCars, getLocations, Car } from '../services/api';
import { CarCard } from '../components/cars/CarCard';
import { AIAssistant } from '../components/ai/AIAssistant';
import { useBookingStore } from '../store/bookingStore';
import { useNavigate } from 'react-router-dom';
import { format, addDays } from 'date-fns';

export const SearchPage: React.FC = () => {
  const navigate = useNavigate();
  const { setSearchParams, setSelectedCar, setStep } = useBookingStore();

  const tomorrow = format(addDays(new Date(), 1), 'yyyy-MM-dd');
  const dayAfter = format(addDays(new Date(), 3), 'yyyy-MM-dd');

  const [location, setLocation] = useState('');
  const [pickupDate, setPickupDate] = useState(tomorrow);
  const [returnDate, setReturnDate] = useState(dayAfter);
  const [category, setCategory] = useState('');
  const [maxRate, setMaxRate] = useState('');
  const [searched, setSearched] = useState(false);
  const [aiHighlights, setAiHighlights] = useState<number[]>([]);

  const { data: locations } = useQuery({
    queryKey: ['locations'],
    queryFn: getLocations,
  });

  const { data: results, isLoading, refetch } = useQuery({
    queryKey: ['cars', location, pickupDate, returnDate, category, maxRate],
    queryFn: () => searchCars({
      location,
      pickup_date: new Date(pickupDate).toISOString(),
      return_date: new Date(returnDate).toISOString(),
      category: category || undefined,
      max_daily_rate: maxRate ? parseFloat(maxRate) : undefined,
    }),
    enabled: false,
  });

  const handleSearch = () => {
    setSearched(true);
    setSearchParams({ location, pickup_date: pickupDate, return_date: returnDate });
    refetch();
  };

  const handleSelectCar = (car: Car) => {
    setSelectedCar(car);
    setStep('confirm');
    navigate(`/book/${car.id}`);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Hero */}
      <div className="bg-gradient-to-r from-blue-700 to-indigo-700 text-white py-14 px-4">
        <div className="max-w-4xl mx-auto">
          <div className="flex justify-end mb-4">
            <button
              onClick={() => navigate('/my-bookings')}
              className="text-sm text-blue-200 hover:text-white border border-blue-400 hover:border-white px-4 py-1.5 rounded-xl transition-colors font-medium"
            >
              📋 My Bookings
            </button>
          </div>
          <div className="text-center">
            <h1 className="text-4xl font-bold mb-2">🚗 DriveAI Car Rentals</h1>
            <p className="text-blue-200 text-lg">Find your perfect car — AI-powered recommendations included</p>
          </div>
        </div>
      </div>

      {/* Search Bar */}
      <div className="max-w-4xl mx-auto -mt-8 px-4">
        <div className="bg-white rounded-2xl shadow-lg p-6">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-4">
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">LOCATION</label>
              <select
                value={location}
                onChange={(e) => setLocation(e.target.value)}
                className="w-full border border-gray-300 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
              >
                <option value="">Any location</option>
                {locations?.map((l) => <option key={l} value={l}>{l}</option>)}
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">PICK-UP DATE</label>
              <input
                type="date"
                value={pickupDate}
                min={tomorrow}
                onChange={(e) => setPickupDate(e.target.value)}
                className="w-full border border-gray-300 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">RETURN DATE</label>
              <input
                type="date"
                value={returnDate}
                min={pickupDate}
                onChange={(e) => setReturnDate(e.target.value)}
                className="w-full border border-gray-300 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
              />
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 mb-1">CATEGORY</label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full border border-gray-300 rounded-xl px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
              >
                <option value="">All</option>
                <option value="economy">Economy</option>
                <option value="standard">Standard</option>
                <option value="luxury">Luxury</option>
                <option value="suv">SUV</option>
                <option value="van">Van</option>
              </select>
            </div>
          </div>
          <button
            onClick={handleSearch}
            className="w-full bg-blue-600 hover:bg-blue-700 text-white font-bold py-3 px-6 rounded-xl transition-colors text-lg"
          >
            Search Available Cars
          </button>
        </div>
      </div>

      {/* Results */}
      <div className="max-w-4xl mx-auto px-4 py-8">
        {searched && (
          <AIAssistant onRecommendations={setAiHighlights} />
        )}

        {isLoading && (
          <div className="text-center py-12 text-gray-500">
            <div className="text-4xl mb-3">🔍</div>
            <p>Searching available cars...</p>
          </div>
        )}

        {results && results.cars.length === 0 && (
          <div className="text-center py-12 text-gray-500">
            <div className="text-4xl mb-3">😔</div>
            <p>No cars available for your selected criteria.</p>
          </div>
        )}

        {results && results.cars.length > 0 && (
          <>
            <p className="text-sm text-gray-500 mb-4">{results.total} car{results.total !== 1 ? 's' : ''} found</p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {results.cars.map((car) => (
                <CarCard
                  key={car.id}
                  car={car}
                  onSelect={handleSelectCar}
                  highlighted={aiHighlights.includes(car.id)}
                />
              ))}
            </div>
          </>
        )}
      </div>
    </div>
  );
};
