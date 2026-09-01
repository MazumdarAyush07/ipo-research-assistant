import { Financials } from "@/lib/api";

export default function FinancialsTable({ financials }: { financials: Financials[] }) {
  if (!financials || financials.length === 0) {
    return <div className="text-gray-400 p-4">Financial data unavailable.</div>;
  }

  // Sort by Year
  const sorted = [...financials].sort((a, b) => {
    // Assuming year is string like "2022", "2023"
    return a.Year.localeCompare(b.Year);
  });

  return (
    <div className="glass-panel overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-gray-300">
          <thead className="bg-black/50 text-xs uppercase text-gray-400">
            <tr>
              <th scope="col" className="px-6 py-4 font-medium">Metric (₹ Cr)</th>
              {sorted.map((f) => (
                <th key={f.Year} scope="col" className="px-6 py-4 font-medium text-right">{f.Year}</th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            <TableRow label="Revenue" data={sorted.map(f => f.Revenue)} />
            <TableRow label="PAT (Profit After Tax)" data={sorted.map(f => f.Pat)} highlight />
            <TableRow label="EBITDA" data={sorted.map(f => f.Ebitda)} />
            <TableRow label="Total Debt" data={sorted.map(f => f.Debt)} />
            <TableRow label="Net Worth" data={sorted.map(f => f.NetWorth)} />
          </tbody>
        </table>
      </div>
    </div>
  );
}

function TableRow({ label, data, highlight = false }: { label: string; data: number[]; highlight?: boolean }) {
  return (
    <tr className="hover:bg-white/5 transition-colors">
      <th scope="row" className={`whitespace-nowrap px-6 py-4 font-medium ${highlight ? "text-indigo-400" : "text-white"}`}>
        {label}
      </th>
      {data.map((val, idx) => (
        <td key={idx} className="px-6 py-4 text-right tabular-nums">
          {val > 0 ? val.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : "-"}
        </td>
      ))}
    </tr>
  );
}
