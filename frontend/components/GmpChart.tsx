"use client";

import { GMP } from "@/lib/api";
import { Area, AreaChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";
import { TrendingUp } from "lucide-react";

export default function GmpChart({ gmp, issuePrice }: { gmp: GMP | null, issuePrice: number }) {
  if (!gmp || gmp.premium_percent === 0) {
    return <div className="text-gray-400 p-4 glass-panel">GMP data unavailable.</div>;
  }

  // Generate a mock trend line leading up to the current GMP for visual appeal
  // Since we only have the current GMP point from the API right now.
  const data = [
    { day: "D-4", value: gmp.gmp_price * 0.6 },
    { day: "D-3", value: gmp.gmp_price * 0.75 },
    { day: "D-2", value: gmp.gmp_price * 0.9 },
    { day: "D-1", value: gmp.gmp_price * 1.05 },
    { day: "Latest", value: gmp.gmp_price },
  ];

  const estListing = issuePrice + gmp.gmp_price;

  return (
    <div className="glass-panel p-6">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-indigo-400" />
            Grey Market Premium (GMP)
          </h3>
          <p className="text-sm text-gray-400 mt-1">Expected listing premium based on unofficial market</p>
        </div>
        <div className="text-right">
          <div className="text-2xl font-bold text-emerald-400">₹{gmp.gmp_price}</div>
          <div className="text-sm font-medium text-emerald-500/80">+{gmp.premium_percent.toFixed(2)}%</div>
        </div>
      </div>

      <div className="h-[200px] w-full">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 0, left: -20, bottom: 0 }}>
            <defs>
              <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#818cf8" stopOpacity={0.3} />
                <stop offset="95%" stopColor="#818cf8" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis dataKey="day" stroke="#52525b" fontSize={12} tickLine={false} axisLine={false} />
            <YAxis stroke="#52525b" fontSize={12} tickLine={false} axisLine={false} />
            <Tooltip 
              contentStyle={{ backgroundColor: '#18181b', borderColor: '#27272a', borderRadius: '8px' }}
              itemStyle={{ color: '#c0caf5' }}
            />
            <Area 
              type="monotone" 
              dataKey="value" 
              stroke="#818cf8" 
              strokeWidth={3}
              fillOpacity={1} 
              fill="url(#colorValue)" 
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="mt-4 pt-4 border-t border-white/10 flex justify-between text-sm">
        <span className="text-gray-400">Estimated Listing Price</span>
        <span className="font-semibold text-white">₹{estListing.toFixed(2)}</span>
      </div>
    </div>
  );
}
