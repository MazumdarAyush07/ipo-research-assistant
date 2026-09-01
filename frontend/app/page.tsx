import { fetchIPOs } from "@/lib/api";
import Link from "next/link";
import { format, isBefore, isAfter } from "date-fns";
import { ChevronRight, Calendar, Building2, TrendingUp, Activity, IndianRupee, Clock } from "lucide-react";

export const revalidate = 60; // Revalidate every 60 seconds

function getStatus(openDateStr: string, closeDateStr: string) {
  if (!openDateStr || !closeDateStr) return { label: "TBA", color: "bg-gray-500/10 text-gray-400 border-gray-500/20" };
  
  const now = new Date();
  const open = new Date(openDateStr);
  const close = new Date(closeDateStr);
  // Add 1 day to close date to account for full day
  close.setHours(23, 59, 59, 999);

  if (isBefore(now, open)) return { label: "Upcoming", color: "bg-amber-500/10 text-amber-400 border-amber-500/20" };
  if (isAfter(now, close)) return { label: "Closed", color: "bg-slate-500/10 text-slate-400 border-slate-500/20" };
  return { label: "Open", color: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20" };
}

function IPOCard({ ipo, status }: { ipo: any, status: any }) {
  return (
    <Link href={`/ipo/${ipo.id}`} key={ipo.id}>
      <div className="relative group overflow-hidden rounded-2xl bg-[#13141b]/80 border border-white/5 hover:border-indigo-500/30 transition-all duration-300 shadow-xl cursor-pointer h-full flex flex-col backdrop-blur-xl">
        <div className="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-indigo-500/0 via-indigo-500/40 to-indigo-500/0 opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
        <div className="absolute -inset-24 bg-gradient-to-br from-indigo-500/10 to-transparent opacity-0 group-hover:opacity-100 blur-2xl transition-all duration-700 pointer-events-none" />
        
        <div className="p-6 flex-1 flex flex-col relative z-10">
          <div className="mb-4 flex items-start justify-between">
            <div>
              <div className="flex items-center gap-2 mb-2">
                <span className="rounded-md bg-white/5 px-2 py-1 text-[10px] font-semibold text-gray-300 uppercase tracking-wider border border-white/10 shadow-sm">
                  {ipo.exchange_type || "IPO"}
                </span>
                <span className={`rounded-md px-2 py-1 text-[10px] font-bold uppercase tracking-wider border ${status.color}`}>
                  {status.label}
                </span>
              </div>
              <h3 className="font-bold text-xl text-white group-hover:text-indigo-400 transition-colors line-clamp-2 leading-tight">
                {ipo.company_name}
              </h3>
              <div className="mt-2 flex items-center gap-1.5 text-xs text-gray-400">
                <Building2 className="h-3.5 w-3.5" />
                <span className="truncate">{ipo.sector || "Unknown Sector"}</span>
              </div>
            </div>
          </div>

          <div className="mt-auto pt-5 grid grid-cols-2 gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-gray-500">
                <Calendar className="h-3.5 w-3.5" />
                Open Date
              </div>
              <div className="font-medium text-sm text-gray-200">
                {ipo.open_date ? format(new Date(ipo.open_date), "MMM dd") : "TBD"}
              </div>
            </div>
            
            <div className="space-y-1">
              <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-gray-500">
                <Clock className="h-3.5 w-3.5" />
                Close Date
              </div>
              <div className="font-medium text-sm text-gray-200">
                {ipo.close_date ? format(new Date(ipo.close_date), "MMM dd") : "TBD"}
              </div>
            </div>

            <div className="space-y-1">
              <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-gray-500">
                <TrendingUp className="h-3.5 w-3.5" />
                Issue Price
              </div>
              <div className="font-semibold text-[15px] text-white">
                ₹{ipo.issue_price > 0 ? ipo.issue_price : "TBD"}
              </div>
            </div>

            <div className="space-y-1">
              <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-gray-500">
                <Calendar className="h-3.5 w-3.5" />
                Listing Date
              </div>
              <div className="font-semibold text-sm text-white">
                {ipo.listing_date ? format(new Date(ipo.listing_date), "MMM dd") : "TBD"}
              </div>
            </div>
          </div>
        </div>

        <div className="h-0.5 w-full bg-gradient-to-r from-transparent via-white/5 to-transparent group-hover:via-indigo-500/20 transition-all duration-300" />
      </div>
    </Link>
  );
}

export default async function Dashboard() {
  const ipos = await fetchIPOs();
  
  const iposWithStatus = ipos.map((ipo) => ({
    ipo,
    status: getStatus(ipo.open_date, ipo.close_date)
  }));

  const activeIpos = iposWithStatus.filter((item) => item.status.label !== "Closed");
  const closedIpos = iposWithStatus.filter((item) => item.status.label === "Closed");

  return (
    <div className="space-y-8 animate-in fade-in slide-in-from-bottom-4 duration-700">
      <div className="flex flex-col gap-2">
        <h1 className="text-3xl font-bold tracking-tight text-white">IPO Dashboard</h1>
        <p className="text-gray-400">Track and analyze recent and upcoming initial public offerings.</p>
      </div>

      {activeIpos.length > 0 && (
        <div className="space-y-4">
          <h2 className="text-xl font-semibold text-white border-b border-white/10 pb-2">Active & Upcoming IPOs</h2>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {activeIpos.map((item) => <IPOCard key={item.ipo.id} ipo={item.ipo} status={item.status} />)}
          </div>
        </div>
      )}

      {closedIpos.length > 0 && (
        <div className="space-y-4 mt-12">
          <h2 className="text-xl font-semibold text-white border-b border-white/10 pb-2">Closed IPOs</h2>
          <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {closedIpos.map((item) => <IPOCard key={item.ipo.id} ipo={item.ipo} status={item.status} />)}
          </div>
        </div>
      )}
      
      {ipos.length === 0 && (
        <div className="flex flex-col items-center justify-center py-20 text-center glass-panel">
          <Activity className="h-10 w-10 text-gray-500 mb-4" />
          <h3 className="text-xl font-medium text-white">No IPOs tracked</h3>
          <p className="mt-2 text-gray-400">Sync data from the backend to populate the dashboard.</p>
        </div>
      )}
    </div>
  );
}
