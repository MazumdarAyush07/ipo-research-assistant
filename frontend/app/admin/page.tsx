"use client";

import { useState, useEffect } from "react";
import { Terminal, Database, Activity, FileText, CheckCircle2, XCircle, RefreshCw, BarChart2, Brain } from "lucide-react";
import { IPO, fetchIPOs } from "@/lib/api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

interface AuditResult {
  total_expected: number;
  downloaded: number;
  unexpected_eof: number;
  below_500kb: number;
  below_50_pages: number;
  missing: string[];
  errors: string[];
}

export default function AdminDashboard() {
  const [ipos, setIpos] = useState<IPO[]>([]);
  const [audit, setAudit] = useState<any>(null);
  const [parsingAudit, setParsingAudit] = useState<any>(null);
  const [analysisAudit, setAnalysisAudit] = useState<any>(null);
  const [trackerAudit, setTrackerAudit] = useState<any>(null);
  const [trackerTimeframe, setTrackerTimeframe] = useState<number>(0);
  const [scoringAudit, setScoringAudit] = useState<any>(null);
  const [reportAudit, setReportAudit] = useState<any>(null);
  const [queues, setQueues] = useState<any[]>([]);
  
  const [isAuditing, setIsAuditing] = useState(false);
  const [isAuditingParsing, setIsAuditingParsing] = useState(false);
  const [isAuditingAnalysis, setIsAuditingAnalysis] = useState(false);
  const [isAuditingTracker, setIsAuditingTracker] = useState(false);
  const [isAuditingScoring, setIsAuditingScoring] = useState(false);
  const [isAuditingReport, setIsAuditingReport] = useState(false);
  
  const [logs, setLogs] = useState<string[]>([]);
  const [isProcessing, setIsProcessing] = useState(false);

  const addLog = (msg: string) => {
    setLogs((prev) => [...prev, `[${new Date().toLocaleTimeString()}] ${msg}`]);
  };

  useEffect(() => {
    fetchInitialData();

    fetchQueues();
    const interval = setInterval(fetchQueues, 30000); // Poll every 30 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchQueues = async () => {
    try {
      const [queueRes, parsingAuditRes, analysisAuditRes, trackerAuditRes, scoringAuditRes, reportAuditRes] = await Promise.all([
        fetch(`${API_BASE_URL}/admin/queues`),
        fetch(`${API_BASE_URL}/admin/parsing-audit`).then(r => r.ok ? r.json() : null),
        fetch(`${API_BASE_URL}/admin/analysis-audit`).then(r => r.ok ? r.json() : null),
        fetch(`${API_BASE_URL}/admin/tracker-audit`).then(r => r.ok ? r.json() : null),
        fetch(`${API_BASE_URL}/admin/scoring-audit`).then(r => r.ok ? r.json() : null),
        fetch(`${API_BASE_URL}/admin/report-audit`).then(r => r.ok ? r.json() : null)
      ]);
      
      if (queueRes.ok) {
        const data = await queueRes.json();
        setQueues(data.data || []);
      }
      if (parsingAuditRes) setParsingAudit(parsingAuditRes.data || parsingAuditRes);
      if (analysisAuditRes) setAnalysisAudit(analysisAuditRes.data || analysisAuditRes);
      if (trackerAuditRes) setTrackerAudit(trackerAuditRes.data || trackerAuditRes);
      if (scoringAuditRes) setScoringAudit(scoringAuditRes.data || scoringAuditRes);
      if (reportAuditRes) setReportAudit(reportAuditRes.data || reportAuditRes);
    } catch (err) {
      console.error("Failed to fetch queue stats", err);
    }
  };

  const fetchInitialData = async () => {
    try {
      const data = await fetchIPOs();
      setIpos(data);
    } catch (err) {
      addLog(`Error fetching initial data: ${err}`);
    }
  };

  const runAudit = async () => {
    if (isAuditing) return;
    setIsAuditing(true);
    addLog("Starting storage audit (this may take up to a minute)...");
    try {
      const res = await fetch(`${API_BASE_URL}/admin/audit`);
      if (res.ok) {
        const data = await res.json();
        setAudit(data.data || data);
        addLog("Storage audit completed.");
      } else {
        addLog(`Storage audit failed: HTTP ${res.status}`);
      }
    } catch (err) {
      addLog(`Error fetching audit data: ${err}`);
    }
    setIsAuditing(false);
  };

  const runParsingAudit = async () => {
    if (isAuditingParsing) return;
    setIsAuditingParsing(true);
    addLog("Starting parsing audit...");
    try {
      const res = await fetch(`${API_BASE_URL}/admin/parsing-audit`);
      if (res.ok) {
        const data = await res.json();
        setParsingAudit(data.data || data);
        addLog("Parsing audit completed.");
      }
    } catch (err) { addLog(`Error fetching parsing audit data: ${err}`); }
    setIsAuditingParsing(false);
  };

  const runAnalysisAudit = async () => {
    if (isAuditingAnalysis) return;
    setIsAuditingAnalysis(true);
    addLog("Starting analysis audit...");
    try {
      const res = await fetch(`${API_BASE_URL}/admin/analysis-audit`);
      if (res.ok) {
        const data = await res.json();
        setAnalysisAudit(data.data || data);
        addLog("Analysis audit completed.");
      }
    } catch (err) { addLog(`Error fetching analysis audit data: ${err}`); }
    setIsAuditingAnalysis(false);
  };

  const runTrackerAudit = async () => {
    if (isAuditingTracker) return;
    setIsAuditingTracker(true);
    addLog(`Starting tracker audit (hours: ${trackerTimeframe})...`);
    try {
      const res = await fetch(`${API_BASE_URL}/admin/tracker-audit?hours=${trackerTimeframe}`);
      if (res.ok) {
        const data = await res.json();
        setTrackerAudit(data.data || data);
        addLog("Tracker audit completed.");
      }
    } catch (err) { addLog(`Error fetching tracker audit data: ${err}`); }
    setIsAuditingTracker(false);
  };

  const runScoringAudit = async () => {
    if (isAuditingScoring) return;
    setIsAuditingScoring(true);
    addLog("Starting scoring audit...");
    try {
      const res = await fetch(`${API_BASE_URL}/admin/scoring-audit`);
      if (res.ok) {
        const data = await res.json();
        setScoringAudit(data.data || data);
        addLog("Scoring audit completed.");
      }
    } catch (err) { addLog(`Error fetching scoring audit data: ${err}`); }
    setIsAuditingScoring(false);
  };

  const runReportAudit = async () => {
    if (isAuditingReport) return;
    setIsAuditingReport(true);
    addLog("Starting report audit...");
    try {
      const res = await fetch(`${API_BASE_URL}/admin/report-audit`);
      if (res.ok) {
        const data = await res.json();
        setReportAudit(data.data || data);
        addLog("Report audit completed.");
      }
    } catch (err) { addLog(`Error fetching report audit data: ${err}`); }
    setIsAuditingReport(false);
  };

  const runBatchJob = async (jobName: string, endpointSuffix: string) => {
    if (isProcessing) return;
    if (ipos.length === 0) {
      addLog(`Cannot run ${jobName}: No IPOs found in database.`);
      return;
    }

    setIsProcessing(true);
    addLog(`Starting batch job: ${jobName} for ${ipos.length} IPOs...`);

    let successCount = 0;
    let failCount = 0;
    let skipCount = 0;

    for (const ipo of ipos) {
      addLog(`Triggering ${jobName} for IPO ${ipo.id} (${ipo.company_name})...`);
      try {
        const res = await fetch(`${API_BASE_URL}/ipos/${ipo.id}/${endpointSuffix}`, { method: "POST" });
        if (res.ok) {
          const resData = await res.json();
          if (resData.status === "skipped") {
            skipCount++;
            addLog(`Skipped IPO ${ipo.id}: ${resData.message}`);
          } else {
            successCount++;
          }
        } else {
          failCount++;
          addLog(`Failed for IPO ${ipo.id}: HTTP ${res.status}`);
        }
      } catch (err) {
        failCount++;
        addLog(`Error for IPO ${ipo.id}: ${err}`);
      }
      
      // Artificial delay to prevent hammering the backend
      await new Promise(r => setTimeout(r, 500));
    }

    addLog(`Finished sending ${jobName} tasks to background queue. Enqueued: ${successCount}, Skipped: ${skipCount}, Failed: ${failCount}`);
    setIsProcessing(false);
  };

  const runTrackersSync = async () => {
    if (isProcessing) return;
    setIsProcessing(true);
    addLog("Triggering global trackers sync...");
    try {
      const res = await fetch(`${API_BASE_URL}/trackers/sync`, { method: "POST" });
      if (res.ok) {
        addLog("Trackers sync triggered successfully.");
      } else {
        addLog(`Failed to trigger trackers sync: HTTP ${res.status}`);
      }
    } catch (err) {
      addLog(`Error triggering trackers sync: ${err}`);
    }
    setIsProcessing(false);
  };

  const runManualIpoSync = async () => {
    if (isProcessing) return;
    setIsProcessing(true);
    addLog("Triggering manual IPO sync...");
    try {
      const res = await fetch(`${API_BASE_URL}/ipos`, { method: "POST" });
      if (res.ok) {
        addLog("Manual IPO sync triggered successfully.");
        // Re-fetch IPOs after a few seconds
        setTimeout(fetchInitialData, 3000);
      } else {
        addLog(`Failed to trigger manual IPO sync: HTTP ${res.status}`);
      }
    } catch (err) {
      addLog(`Error triggering manual IPO sync: ${err}`);
    }
    setIsProcessing(false);
  };

  return (
    <div className="min-h-screen bg-[#0A0E17] p-8 font-sans">
      <div className="max-w-6xl mx-auto space-y-8">
        
        {/* Header */}
        <div className="flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-white tracking-tight">Pipeline Admin</h1>
            <p className="text-gray-400 mt-2">Manage backend tasks, ingestions, and audits.</p>
          </div>
          <div className="flex gap-4">
            <button 
              onClick={fetchInitialData}
              className="px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg flex items-center gap-2 border border-white/10 transition-colors"
            >
              <RefreshCw className="h-4 w-4" />
              Refresh Status
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Controls Column */}
          <div className="lg:col-span-1 space-y-6">
            
            <div className="glass-panel p-6 rounded-xl space-y-6">
              <h2 className="text-xl font-semibold text-white flex items-center gap-2 border-b border-white/10 pb-4">
                <Activity className="h-5 w-5 text-indigo-400" />
                Actions
              </h2>

              <div className="space-y-3">
                <button 
                  onClick={runManualIpoSync}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-blue-600/20 hover:bg-blue-600/40 border border-blue-500/30 text-blue-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><Database className="h-4 w-4" /> 1. Sync IPO Calendar</span>
                </button>

                <button 
                  onClick={() => runBatchJob("Batch Ingest (Downloads)", "documents/trigger")}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-purple-600/20 hover:bg-purple-600/40 border border-purple-500/30 text-purple-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><FileText className="h-4 w-4" /> 2. Download DRHPs</span>
                </button>

                <button 
                  onClick={() => runBatchJob("Batch Parse (AI)", "parse/trigger")}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-emerald-600/20 hover:bg-emerald-600/40 border border-emerald-500/30 text-emerald-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><CheckCircle2 className="h-4 w-4" /> 3. Parse PDFs (AI)</span>
                </button>

                <button 
                  onClick={() => runBatchJob("Run AI Analyst", "analysis/trigger")}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-indigo-600/20 hover:bg-indigo-600/40 border border-indigo-500/30 text-indigo-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><Brain className="h-4 w-4" /> 4. Run AI Analyst</span>
                </button>

                <button 
                  onClick={runTrackersSync}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-amber-600/20 hover:bg-amber-600/40 border border-amber-500/30 text-amber-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><BarChart2 className="h-4 w-4" /> 5. Sync Trackers (Live)</span>
                </button>

                <button 
                  onClick={() => runBatchJob("Batch Score", "score/trigger")}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-rose-600/20 hover:bg-rose-600/40 border border-rose-500/30 text-rose-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><Activity className="h-4 w-4" /> 6. Run Scoring Engine</span>
                </button>
                
                <button 
                  onClick={() => runBatchJob("Generate HTML Reports", "report/generate")}
                  disabled={isProcessing}
                  className="w-full px-4 py-3 bg-slate-600/20 hover:bg-slate-600/40 border border-slate-500/30 text-slate-200 rounded-lg flex items-center justify-between transition-colors disabled:opacity-50"
                >
                  <span className="flex items-center gap-2"><FileText className="h-4 w-4" /> 7. Generate Reports</span>
                </button>
              </div>
            </div>

                        {/* Background Tasks Monitor */}
            <div className="glass-panel p-6 rounded-xl mt-6">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <Activity className="h-5 w-5 text-indigo-400" />
                  Background Tasks
                </div>
                <div className="flex items-center gap-3">
                  {queues.some(q => q.active > 0 || q.pending > 0) && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-indigo-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-indigo-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={fetchQueues}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white"
                    title="Refresh Queue Stats"
                  >
                    <RefreshCw className="h-4 w-4" />
                  </button>
                </div>
              </h2>
              {queues.length === 0 ? (
                <div className="text-gray-500 text-sm text-center py-4">No active queues found.</div>
              ) : (
                <div className="space-y-4">
                  {queues.map((q, idx) => (
                    <div key={idx} className="bg-white/5 p-4 rounded-lg border border-white/10">
                      <div className="flex justify-between items-center mb-3">
                        <span className="font-medium text-white">{q.queue}</span>
                        <span className="text-xs bg-indigo-500/20 text-indigo-300 px-2 py-1 rounded">
                          {q.active + q.pending} in queue
                        </span>
                      </div>
                      <div className="grid grid-cols-3 gap-2 text-center text-sm">
                        <div className="bg-emerald-500/10 rounded py-2 border border-emerald-500/20">
                          <div className="text-emerald-400 font-bold">{q.active}</div>
                          <div className="text-emerald-500/50 text-[10px] uppercase tracking-wider">Active</div>
                        </div>
                        <div className="bg-amber-500/10 rounded py-2 border border-amber-500/20">
                          <div className="text-amber-400 font-bold">{q.pending}</div>
                          <div className="text-amber-500/50 text-[10px] uppercase tracking-wider">Pending</div>
                        </div>
                        <div className="bg-gray-500/10 rounded py-2 border border-gray-500/20">
                          <div className="text-gray-400 font-bold">{q.completed}</div>
                          <div className="text-gray-500/50 text-[10px] uppercase tracking-wider">Completed</div>
                        </div>
                      </div>
                      {q.retry > 0 && (
                        <div className="mt-2 text-xs text-rose-400 flex items-center gap-1">
                          <XCircle className="h-3 w-3" /> {q.retry} tasks retrying
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* Console Column */}
          <div className="lg:col-span-2">
            <div className="glass-panel rounded-xl h-[600px] flex flex-col border border-white/10 overflow-hidden">
              <div className="bg-black/40 px-4 py-3 border-b border-white/10 flex items-center gap-2">
                <Terminal className="h-4 w-4 text-gray-400" />
                <span className="text-sm font-medium text-gray-300 tracking-wide uppercase">Pipeline Console</span>
                {isProcessing && <span className="ml-auto flex items-center gap-2 text-xs text-indigo-400"><RefreshCw className="h-3 w-3 animate-spin" /> Processing</span>}
              </div>
              <div className="flex-1 p-4 overflow-y-auto bg-black/60 font-mono text-sm">
                {logs.length === 0 ? (
                  <div className="text-gray-600 h-full flex items-center justify-center">
                    Ready to execute pipeline tasks.
                  </div>
                ) : (
                  <div className="space-y-2">
                    {logs.map((log, i) => (
                      <div key={i} className={`${log.includes("Error") || log.includes("Failed") ? "text-rose-400" : log.includes("Completed") || log.includes("Success") ? "text-emerald-400" : "text-gray-300"}`}>
                        {log}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Pipeline Audits */}
        <div className="mt-8 space-y-6">
          <h2 className="text-xl font-semibold text-white tracking-tight flex items-center gap-2 border-b border-white/10 pb-4">
             Pipeline Audits
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {/* Audit Summary */}
            <div className="glass-panel p-6 rounded-xl">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <Database className="h-5 w-5 text-emerald-400" />
                  Storage Audit
                </div>
                <div className="flex items-center gap-3">
                  {isAuditing && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-emerald-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-emerald-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={runAudit}
                    disabled={isAuditing}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white disabled:opacity-50"
                    title="Refresh Storage Audit"
                  >
                    <RefreshCw className={`h-4 w-4 ${isAuditing ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </h2>
              {audit ? (
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Total IPOs Tracked</span>
                    <span className="text-white font-medium">{ipos.length}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">PDFs Downloaded</span>
                    <span className="text-emerald-400 font-medium">{audit.downloaded} / {audit.total_expected}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Files &lt; 50 Pages</span>
                    <span className={audit.below_50_pages > 0 ? "text-rose-400 font-medium" : "text-emerald-400"}>{audit.below_50_pages}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Missing/Corrupt</span>
                    <span className={(audit.missing.length > 0 || audit.unexpected_eof > 0) ? "text-rose-400 font-medium" : "text-emerald-400"}>
                      {audit.missing.length + audit.unexpected_eof}
                    </span>
                  </div>
                  {(audit?.errors?.length > 0 || audit?.missing?.length > 0) && (() => {
                    const allWarnings = [
                      ...(audit.errors || []),
                      ...(audit.missing || []).map((m: any) => `${m}: Missing PDF`)
                    ];
                    return (
                      <div className="mt-4 pt-4 border-t border-white/10">
                        <p className="text-rose-400 font-medium mb-2 flex items-center gap-2 py-1">
                          <XCircle className="h-4 w-4" /> Audit Warnings
                        </p>
                        <ul className="list-disc pl-4 space-y-1 text-gray-400">
                          {allWarnings.slice(0, 50).map((e, i) => <li key={i}>{e}</li>)}
                          {allWarnings.length > 50 && <li>...and {allWarnings.length - 50} more</li>}
                        </ul>
                      </div>
                    );
                  })()}
                </div>
              ) : (
                <div className="text-gray-500 text-sm flex flex-col items-center justify-center py-8 gap-4">
                  {isAuditing ? (
                    <>
                      <RefreshCw className="h-6 w-6 animate-spin text-emerald-400" />
                      <span>Scanning PDFs on disk...</span>
                    </>
                  ) : (
                    <>
                      <p>Audit data is not loaded.</p>
                      <button 
                        onClick={runAudit}
                        className="px-4 py-2 bg-emerald-600/20 hover:bg-emerald-600/40 border border-emerald-500/30 text-emerald-200 rounded-lg flex items-center gap-2 transition-colors"
                      >
                        <Database className="h-4 w-4" /> Run Storage Audit
                      </button>
                    </>
                  )}
                </div>
              )}
            </div>

            {/* Parsing Data Audit */}
            <div className="glass-panel p-6 rounded-xl">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <CheckCircle2 className="h-5 w-5 text-indigo-400" />
                  Parsing Audit
                </div>
                <div className="flex items-center gap-3">
                  {isAuditingParsing && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-indigo-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-indigo-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={runParsingAudit}
                    disabled={isAuditingParsing}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white disabled:opacity-50"
                    title="Refresh Parsing Audit"
                  >
                    <RefreshCw className={`h-4 w-4 ${isAuditingParsing ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </h2>
              {parsingAudit ? (
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Total IPOs Tracked</span>
                    <span className="text-white font-medium">{parsingAudit.total_ipos}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Financials Parsed</span>
                    <span className="text-emerald-400 font-medium">{parsingAudit.total_financials}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Missing Parsing</span>
                    <span className={parsingAudit.missing_parsing > 0 ? "text-rose-400 font-medium" : "text-emerald-400"}>
                      {parsingAudit.missing_parsing}
                    </span>
                  </div>
                </div>
              ) : (
                <div className="text-gray-500 text-sm text-center py-4">Data not loaded.</div>
              )}
            </div>

            {/* Analysis Audit */}
            <div className="glass-panel p-6 rounded-xl">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <Brain className="h-5 w-5 text-purple-400" />
                  Analysis Audit
                </div>
                <div className="flex items-center gap-3">
                  {isAuditingAnalysis && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-purple-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-purple-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={runAnalysisAudit}
                    disabled={isAuditingAnalysis}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white disabled:opacity-50"
                    title="Refresh Analysis Audit"
                  >
                    <RefreshCw className={`h-4 w-4 ${isAuditingAnalysis ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </h2>
              {analysisAudit ? (
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Total IPOs Tracked</span>
                    <span className="text-white font-medium">{analysisAudit.total_ipos}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Analysis Completed</span>
                    <span className="text-emerald-400 font-medium">{analysisAudit.completed}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Missing Analysis</span>
                    <span className={analysisAudit.missing > 0 ? "text-rose-400 font-medium" : "text-emerald-400"}>
                      {analysisAudit.missing}
                    </span>
                  </div>
                </div>
              ) : (
                <div className="text-gray-500 text-sm text-center py-4">Data not loaded.</div>
              )}
            </div>

            {/* Tracker Audit */}
            <div className="glass-panel p-6 rounded-xl">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <BarChart2 className="h-5 w-5 text-amber-400" />
                  Tracker Audit
                </div>
                <div className="flex items-center gap-3">
                  <select 
                    value={trackerTimeframe} 
                    onChange={(e) => { setTrackerTimeframe(Number(e.target.value)); }}
                    className="bg-white/5 border border-white/10 rounded-md text-sm text-gray-300 px-2 py-1 outline-none"
                  >
                    <option value={3}>Last 3h</option>
                    <option value={6}>Last 6h</option>
                    <option value={24}>Last 24h</option>
                    <option value={0}>All Time</option>
                  </select>
                  {isAuditingTracker && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-amber-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-amber-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={runTrackerAudit}
                    disabled={isAuditingTracker}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white disabled:opacity-50"
                    title="Refresh Tracker Audit"
                  >
                    <RefreshCw className={`h-4 w-4 ${isAuditingTracker ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </h2>
              {trackerAudit ? (
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Peers Tracked</span>
                    <span className="text-white font-medium">{trackerAudit.peers_tracked}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">GMP Tracked</span>
                    <span className="text-white font-medium">{trackerAudit.gmp_tracked}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Subscriptions Tracked</span>
                    <span className="text-white font-medium">{trackerAudit.subscriptions_tracked}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2">
                    <span className="text-gray-400">Valuations Tracked</span>
                    <span className="text-white font-medium">{trackerAudit.valuations_tracked}</span>
                  </div>
                </div>
              ) : (
                <div className="text-gray-500 text-sm text-center py-4">Click refresh to load audit data</div>
              )}
            </div>

            {/* Scoring Audit */}
            <div className="glass-panel p-6 rounded-xl">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <Activity className="h-5 w-5 text-rose-400" />
                  Scoring Audit
                </div>
                <div className="flex items-center gap-3">
                  {isAuditingScoring && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-rose-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-rose-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={runScoringAudit}
                    disabled={isAuditingScoring}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white disabled:opacity-50"
                    title="Refresh Scoring Audit"
                  >
                    <RefreshCw className={`h-4 w-4 ${isAuditingScoring ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </h2>
              {scoringAudit ? (
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Total IPOs Tracked</span>
                    <span className="text-white font-medium">{scoringAudit.total_ipos}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Scores Generated</span>
                    <span className="text-emerald-400 font-medium">{scoringAudit.completed}</span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-400">Missing Scores</span>
                    <span className={scoringAudit.missing > 0 ? "text-rose-400 font-medium" : "text-emerald-400"}>
                      {scoringAudit.missing}
                    </span>
                  </div>
                </div>
              ) : (
                <div className="text-gray-500 text-sm text-center py-4">Data not loaded.</div>
              )}
            </div>

            {/* Report Audit */}
            <div className="glass-panel p-6 rounded-xl">
              <h2 className="text-xl font-semibold text-white flex items-center justify-between border-b border-white/10 pb-4 mb-4">
                <div className="flex items-center gap-2">
                  <FileText className="h-5 w-5 text-slate-400" />
                  Report Audit
                </div>
                <div className="flex items-center gap-3">
                  {isAuditingReport && (
                    <span className="flex h-3 w-3">
                      <span className="animate-ping absolute inline-flex h-3 w-3 rounded-full bg-slate-400 opacity-75"></span>
                      <span className="relative inline-flex rounded-full h-3 w-3 bg-slate-500"></span>
                    </span>
                  )}
                  <button 
                    onClick={runReportAudit}
                    disabled={isAuditingReport}
                    className="p-1.5 bg-white/5 hover:bg-white/10 rounded-md transition-colors text-gray-400 hover:text-white disabled:opacity-50"
                    title="Refresh Report Audit"
                  >
                    <RefreshCw className={`h-4 w-4 ${isAuditingReport ? 'animate-spin' : ''}`} />
                  </button>
                </div>
              </h2>
              {reportAudit ? (
                <div className="space-y-4 text-sm">
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Total IPOs Tracked</span>
                    <span className="text-white font-medium">{reportAudit.total_ipos}</span>
                  </div>
                  <div className="flex justify-between items-center pb-2 border-b border-white/5">
                    <span className="text-gray-400">Reports Generated</span>
                    <span className="text-emerald-400 font-medium">{reportAudit.completed}</span>
                  </div>
                  <div className="flex flex-col gap-2">
                    <div className="flex justify-between items-center">
                      <span className="text-gray-400">Missing Reports</span>
                      <span className={reportAudit.missing > 0 ? "text-rose-400 font-medium" : "text-emerald-400"}>
                        {reportAudit.missing}
                      </span>
                    </div>
                    {reportAudit.missing > 0 && reportAudit.missing_names && (
                      <div className="text-xs text-rose-300/80 bg-rose-500/10 p-2 rounded-md max-h-32 overflow-y-auto mt-1">
                        <span className="block mb-1 font-medium text-rose-300 border-b border-rose-500/20 pb-1">Missing:</span>
                        <ul className="list-disc list-inside space-y-1">
                          {reportAudit.missing_names.map((name: string, i: number) => (
                            <li key={i} className="truncate" title={name}>{name}</li>
                          ))}
                        </ul>
                      </div>
                    )}
                  </div>
                </div>
              ) : (
                <div className="text-gray-500 text-sm text-center py-4">Data not loaded.</div>
              )}
            </div>

                      </div>
        </div>

      </div>
    </div>
  );
}