import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Link from "next/link";
import { Activity } from "lucide-react";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "IPO Research Assistant",
  description: "AI-powered IPO analysis dashboard",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark">
      <body className={`${inter.className} min-h-screen bg-black text-white selection:bg-indigo-500/30`}>
        
        {/* Abstract Background Gradient */}
        <div className="fixed inset-0 z-[-1] bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-indigo-900/20 via-black to-black"></div>

        {/* Navigation Bar */}
        <nav className="sticky top-0 z-50 w-full border-b border-white/10 bg-black/50 backdrop-blur-md">
          <div className="container mx-auto flex h-16 max-w-7xl items-center px-4 md:px-8">
            <Link href="/" className="flex items-center gap-2 transition-opacity hover:opacity-80">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-500">
                <Activity className="h-5 w-5 text-white" />
              </div>
              <span className="text-xl font-semibold tracking-tight">IPO Assistant</span>
            </Link>
            
            <div className="ml-auto flex items-center space-x-6 text-sm font-medium">
              <Link href="/" className="text-gray-300 hover:text-white transition-colors">
                Dashboard
              </Link>
              <Link href="/history" className="text-gray-300 hover:text-white transition-colors">
                History
              </Link>
            </div>
          </div>
        </nav>

        <main className="container mx-auto max-w-7xl px-4 py-8 md:px-8">
          {children}
        </main>
      </body>
    </html>
  );
}
