import { fetchIPODetails, fetchIPOs } from "@/lib/api";
import { notFound } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, ExternalLink, AlertTriangle, Lightbulb, ShieldAlert } from "lucide-react";
import ScoreBreakdown from "@/components/ScoreBreakdown";
import FinancialsTable from "@/components/FinancialsTable";
import PeerComparison from "@/components/PeerComparison";
import GmpChart from "@/components/GmpChart";

export const revalidate = 60;

export default async function IPODetails({ params }: { params: { id: string } }) {
  // We need basic IPO info for the header. Since fetchIPODetails doesn't return the base IPO struct currently,
  // we can fetch the list and find it, or assume the backend has an endpoint. For now, fetchIPOs is cached.
  const ipos = await fetchIPOs();
  const ipo = ipos.find((i) => i.id.toString() === params.id);
  
  if (!ipo) {
    notFound();
  }

  const { score, financials, peers, ipoPe, gmp, subscriptions, analysis } = await fetchIPODetails(params.id);

  console.log("DEBUG FULL IPO:", JSON.stringify(ipo));
  console.log("DEBUG FULL SCORE:", JSON.stringify(score));
  console.log("DEBUG FULL ANALYSIS:", JSON.stringify(analysis));
  console.log("DEBUG FULL GMP:", JSON.stringify(gmp));
  console.log("DEBUG FULL SUBS:", JSON.stringify(subscriptions));
  console.log("DEBUG FULL FINANCIALS:", JSON.stringify(financials));
  console.log("DEBUG FULL PEERS:", JSON.stringify(peers));

  return (
    <div className="space-y-10 animate-in fade-in slide-in-from-bottom-4 duration-700 pb-20">
      {/* Header */}
      <div className="flex flex-col gap-6 md:flex-row md:items-start md:justify-between">
        <div>
          <Link href="/" className="inline-flex items-center gap-2 text-sm text-gray-400 hover:text-white transition-colors mb-4">
            <ArrowLeft className="h-4 w-4" />
            Back to Dashboard
          </Link>
          <h1 className="text-4xl font-bold tracking-tight text-white">{ipo.company_name}</h1>
          <div className="mt-2 flex items-center gap-3 text-sm text-gray-400">
            <span className="rounded-md bg-white/10 px-2 py-1 font-medium text-gray-200">
              {ipo.exchange_type || "IPO"}
            </span>
            <span>•</span>
            <span>{ipo.sector}</span>
            <span>•</span>
            <span>Issue Price: ₹{ipo.issue_price}</span>
          </div>
        </div>
        
        <div className="flex items-center gap-3">
          <a 
            href={`http://localhost:8080/api/ipos/${ipo.id}/report`}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-indigo-600 px-6 py-2.5 text-sm font-medium text-white transition-colors hover:bg-indigo-700 hover:shadow-[0_0_20px_rgba(79,70,229,0.3)]"
          >
            View Full HTML Report
            <ExternalLink className="h-4 w-4" />
          </a>
        </div>
      </div>

      {/* Score Section */}
      <section>
        <h2 className="text-2xl font-semibold text-white mb-6">AI Evaluation Score</h2>
        <ScoreBreakdown score={score} />
      </section>

      {/* Market Demand (GMP & Subs) */}
      <section className="grid gap-6 lg:grid-cols-2">
        <div>
          <h2 className="text-2xl font-semibold text-white mb-6">Grey Market Premium</h2>
          <GmpChart gmp={gmp} issuePrice={ipo.issue_price} />
        </div>
        <div>
          <h2 className="text-2xl font-semibold text-white mb-6">Live Subscription</h2>
          <div className="glass-panel p-6 h-[282px] flex flex-col justify-center">
            {subscriptions && subscriptions.total_demand > 0 ? (
              <div className="space-y-6">
                <SubsBar label="QIB (Institutions)" value={subscriptions.qib_demand} color="bg-indigo-500" />
                <SubsBar label="NII (HNI)" value={subscriptions.nii_demand} color="bg-emerald-500" />
                <SubsBar label="Retail" value={subscriptions.retail_demand} color="bg-blue-500" />
                <div className="pt-4 border-t border-white/10 flex justify-between items-center">
                  <span className="text-gray-400">Total Subscription</span>
                  <span className="text-2xl font-bold text-white">{subscriptions.total_demand.toFixed(2)}x</span>
                </div>
              </div>
            ) : (
              <div className="text-center text-gray-400">Subscription data unavailable or window not open.</div>
            )}
          </div>
        </div>
      </section>

      {/* Financials & Peers */}
      <section>
        <h2 className="text-2xl font-semibold text-white mb-6">Financial Highlights</h2>
        <FinancialsTable financials={financials} />
      </section>

      <section>
        <h2 className="text-2xl font-semibold text-white mb-6">Peer Comparison</h2>
        <PeerComparison peers={peers} ipoName={ipo.company_name} ipoPe={ipoPe} />
      </section>

      {/* AI Analyst Findings */}
      {analysis && (
        <section>
          <h2 className="text-2xl font-semibold text-white mb-6">AI Analyst Findings</h2>
          <div className="grid gap-6 md:grid-cols-3">
            <div className="glass-panel p-6 border-t-2 border-t-rose-500">
              <div className="flex items-center gap-2 mb-4">
                <AlertTriangle className="h-5 w-5 text-rose-500" />
                <h3 className="font-semibold text-white">Critical Red Flags</h3>
              </div>
              <ul className="text-sm text-gray-300 leading-relaxed space-y-3">
                {analysis.red_flags.split(/\s*(?:\d+\.)\s+/).filter(Boolean).map((flag, idx) => (
                  <li key={idx} className="flex items-start gap-2">
                    <span className="text-rose-500 font-bold mt-0.5">•</span>
                    <span>{flag.trim()}</span>
                  </li>
                ))}
              </ul>
            </div>
            
            <div className="glass-panel p-6 border-t-2 border-t-amber-500">
              <div className="flex items-center gap-2 mb-4">
                <ShieldAlert className="h-5 w-5 text-amber-500" />
                <h3 className="font-semibold text-white">Unique Risks</h3>
              </div>
              <p className="text-sm text-gray-300 whitespace-pre-wrap leading-relaxed">{analysis.unique_risks}</p>
            </div>
            
            <div className="glass-panel p-6 border-t-2 border-t-indigo-500">
              <div className="flex items-center gap-2 mb-4">
                <Lightbulb className="h-5 w-5 text-indigo-500" />
                <h3 className="font-semibold text-white">Management Assumptions</h3>
              </div>
              <p className="text-sm text-gray-300 whitespace-pre-wrap leading-relaxed">{analysis.management_assumptions}</p>
            </div>
          </div>
        </section>
      )}
    </div>
  );
}

function SubsBar({ label, value, color }: { label: string; value: number; color: string }) {
  // Cap at 100x for visual scale
  const percent = Math.min((value / 100) * 100, 100);
  return (
    <div>
      <div className="flex justify-between text-sm mb-1.5">
        <span className="text-gray-300">{label}</span>
        <span className="font-semibold text-white">{value.toFixed(2)}x</span>
      </div>
      <div className="h-2 w-full rounded-full bg-black/50 overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${percent}%` }} />
      </div>
    </div>
  );
}
