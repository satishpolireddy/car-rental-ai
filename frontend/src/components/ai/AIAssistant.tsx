import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getAIRecommendation, AIResponse } from '../../services/api';

interface AIAssistantProps {
  onRecommendations: (carIds: number[]) => void;
}

const suggestions = [
  'I need a family car for a road trip to the mountains',
  'Best electric car for city driving under $100/day',
  'Luxury car for a business trip — need comfort and space',
  'Cheapest option for a weekend getaway',
];

export const AIAssistant: React.FC<AIAssistantProps> = ({ onRecommendations }) => {
  const [query, setQuery] = useState('');
  const [submittedQuery, setSubmittedQuery] = useState('');
  const [response, setResponse] = useState<AIResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (q: string) => {
    if (!q.trim()) return;
    setLoading(true);
    setError('');
    setSubmittedQuery(q);
    try {
      const res = await getAIRecommendation(q);
      setResponse(res);
      if (res.recommended_car_ids?.length > 0) {
        onRecommendations(res.recommended_car_ids);
      }
    } catch {
      setError('AI assistant is unavailable. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="bg-gradient-to-br from-blue-50 to-indigo-50 border border-blue-200 rounded-2xl p-6 mb-6">
      <div className="flex items-center gap-2 mb-4">
        <span className="text-2xl">🤖</span>
        <div>
          <h3 className="font-bold text-gray-900">AI Car Assistant</h3>
          <p className="text-sm text-gray-500">Powered by Azure OpenAI</p>
        </div>
      </div>

      <div className="flex gap-2 mb-4">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          onKeyDown={(e) => e.key === 'Enter' && handleSubmit(query)}
          placeholder="Describe your ideal car or trip..."
          className="flex-1 border border-gray-300 rounded-xl px-4 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-400"
        />
        <button
          onClick={() => handleSubmit(query)}
          disabled={loading}
          className="bg-blue-600 hover:bg-blue-700 disabled:bg-blue-300 text-white px-4 py-2 rounded-xl text-sm font-semibold transition-colors"
        >
          {loading ? '...' : 'Ask'}
        </button>
      </div>

      {/* Quick suggestions */}
      {!response && (
        <div className="flex flex-wrap gap-2 mb-3">
          {suggestions.map((s) => (
            <button
              key={s}
              onClick={() => { setQuery(s); handleSubmit(s); }}
              className="text-xs bg-white border border-blue-200 text-blue-700 px-3 py-1 rounded-full hover:bg-blue-50 transition-colors"
            >
              {s}
            </button>
          ))}
        </div>
      )}

      {error && <p className="text-red-600 text-sm">{error}</p>}

      {response && (
        <div className="bg-white rounded-xl p-4 border border-blue-100">
          <p className="text-sm text-gray-800 leading-relaxed">{response.message}</p>
          {response.recommended_car_ids?.length > 0 && (
            <p className="mt-2 text-xs text-blue-600 font-medium">
              ✨ {response.recommended_car_ids.length} car{response.recommended_car_ids.length > 1 ? 's' : ''} highlighted below
            </p>
          )}
          <p className="mt-2 text-xs text-gray-400">
            {response.latency_ms}ms · {response.tokens_used} tokens
          </p>
          <button
            onClick={() => { setResponse(null); setQuery(''); }}
            className="mt-2 text-xs text-gray-400 hover:text-gray-600 underline"
          >
            Ask another question
          </button>
        </div>
      )}
    </div>
  );
};
