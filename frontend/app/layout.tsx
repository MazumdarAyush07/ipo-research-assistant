import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "IPO Research Assistant",
  description: "AI-powered IPO analysis and intelligence engine",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
