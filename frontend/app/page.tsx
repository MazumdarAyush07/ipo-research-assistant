import { fetchIPOs } from "@/lib/api";
import Link from "next/link";
import { format, isBefore, isAfter } from "date-fns";
import { ChevronRight, Calendar, Building2, TrendingUp, Activity, IndianRupee, Clock } from "lucide-react";

import { IPOClientList } from "@/components/IPOClientList";

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

export default async function Dashboard() {
  const ipos = await fetchIPOs();
  
  const iposWithStatus = ipos.map((ipo) => ({
    ipo,
    status: getStatus(ipo.open_date, ipo.close_date)
  }));

  return <IPOClientList iposWithStatus={iposWithStatus} />;
}
