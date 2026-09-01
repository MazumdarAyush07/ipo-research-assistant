import { Peer } from "@/lib/api";

export default function PeerComparison({ peers, ipoName, ipoPe }: { peers: Peer[], ipoName?: string, ipoPe?: number | null }) {
  if (!peers || peers.length === 0) {
    return <div className="text-gray-400 p-4">Peer comparison data unavailable.</div>;
  }

  return (
    <div className="glass-panel overflow-hidden">
      <div className="overflow-x-auto">
        <table className="w-full text-left text-sm text-gray-300">
          <thead className="bg-black/50 text-xs uppercase text-gray-400">
            <tr>
              <th scope="col" className="px-6 py-4 font-medium">Company</th>
              <th scope="col" className="px-6 py-4 font-medium text-right">P/E Ratio</th>
              <th scope="col" className="px-6 py-4 font-medium text-right">P/B Ratio</th>
              <th scope="col" className="px-6 py-4 font-medium text-right">Market Cap (Cr)</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/5">
            {ipoName && ipoPe != null && (
              <tr className="bg-blue-900/20 hover:bg-blue-900/30 transition-colors relative">
                <th scope="row" className="whitespace-nowrap px-6 py-4 font-medium text-white flex items-center gap-2">
                  <div className="w-2 h-2 rounded-full bg-blue-500"></div>
                  {ipoName}
                </th>
                <td className="px-6 py-4 text-right tabular-nums text-emerald-400 font-semibold">{ipoPe > 0 ? ipoPe.toFixed(2) : "N/A"}</td>
                <td className="px-6 py-4 text-right tabular-nums">-</td>
                <td className="px-6 py-4 text-right tabular-nums">-</td>
              </tr>
            )}
            {peers.map((peer, idx) => (
              <tr key={idx} className="hover:bg-white/5 transition-colors">
                <th scope="row" className="whitespace-nowrap px-6 py-4 font-medium text-white">
                  {peer.name}
                  <span className="ml-2 rounded-md bg-white/10 px-2 py-0.5 text-xs font-normal text-gray-400">
                    {peer.ticker}
                  </span>
                </th>
                <td className="px-6 py-4 text-right tabular-nums text-indigo-400 font-semibold">{peer.pe > 0 ? peer.pe.toFixed(2) : "N/A"}</td>
                <td className="px-6 py-4 text-right tabular-nums">{peer.pb > 0 ? peer.pb.toFixed(2) : "N/A"}</td>
                <td className="px-6 py-4 text-right tabular-nums">
                  {peer.market_cap > 0 ? peer.market_cap.toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : "N/A"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
