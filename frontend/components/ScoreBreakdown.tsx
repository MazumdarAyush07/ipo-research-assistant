"use client";

import { Score } from "@/lib/api";
import { cn } from "@/lib/utils";

export default function ScoreBreakdown({ score }: { score: Score | null }) {
  if (!score) return <div className="text-gray-400">Score data unavailable.</div>;

  const totalScore = score.TotalScore;
  const isGood = totalScore > 60;
  const isBad = totalScore < 40;
  const colorClass = isGood ? "text-emerald-400" : isBad ? "text-rose-400" : "text-amber-400";
  const strokeClass = isGood ? "stroke-emerald-400" : isBad ? "stroke-rose-400" : "stroke-amber-400";

  return (
    <div className="grid gap-6 md:grid-cols-[300px_1fr]">
      {/* Circle Chart for Total Score */}
      <div className="glass-panel flex flex-col items-center justify-center p-8">
        <div className="relative flex items-center justify-center">
          <svg className="h-48 w-48 -rotate-90 transform">
            <circle
              className="stroke-white/10"
              strokeWidth="12"
              fill="transparent"
              r="74"
              cx="96"
              cy="96"
            />
            <circle
              className={cn("transition-all duration-1000 ease-out", strokeClass)}
              strokeWidth="12"
              strokeDasharray={465}
              strokeDashoffset={465 - (465 * totalScore) / 100}
              strokeLinecap="round"
              fill="transparent"
              r="74"
              cx="96"
              cy="96"
            />
          </svg>
          <div className="absolute flex flex-col items-center justify-center">
            <span className={cn("text-5xl font-bold", colorClass)}>{totalScore}</span>
            <span className="text-sm text-gray-400">out of 100</span>
          </div>
        </div>
        <h3 className="mt-6 text-xl font-semibold text-white">{score.Recommendation}</h3>
      </div>

      {/* Grid of sub-scores */}
      <div className="grid gap-4 sm:grid-cols-2">
        <ScoreCard title="Financials" score={score.FinancialsScore} max={40} reason={score.FinancialsReason} />
        <ScoreCard title="Valuation" score={score.ValuationScore} max={20} reason={score.ValuationReason} />
        <ScoreCard title="Promoter" score={score.PromoterScore} max={10} reason={score.PromoterReason} />
        <ScoreCard title="Industry" score={score.IndustryScore} max={10} reason={score.IndustryReason} />
        <ScoreCard title="Risk" score={score.RiskScore} max={10} reason={score.RiskReason} />
        <ScoreCard title="Market Demand" score={score.SubscriptionScore + score.GmpScore} max={10} reason="Combined Subscription & GMP outlook." />
      </div>
    </div>
  );
}

function ScoreCard({ title, score, max, reason }: { title: string; score: number; max: number; reason: string }) {
  const percent = (score / max) * 100;
  const isGood = percent > 60;
  const isBad = percent < 40;
  const barClass = isGood ? "bg-emerald-500" : isBad ? "bg-rose-500" : "bg-amber-500";

  return (
    <div className="glass-panel flex flex-col justify-between p-5 hover:bg-white/5 transition-colors">
      <div className="flex items-center justify-between mb-2">
        <h4 className="font-medium text-white">{title}</h4>
        <span className="text-sm font-semibold text-gray-300">
          {score} <span className="text-gray-500">/ {max}</span>
        </span>
      </div>
      <div className="h-1.5 w-full rounded-full bg-black/50 mb-4 overflow-hidden">
        <div className={cn("h-full rounded-full", barClass)} style={{ width: `${percent}%` }} />
      </div>
      <p className="text-sm text-gray-400 leading-relaxed">{reason}</p>
    </div>
  );
}
