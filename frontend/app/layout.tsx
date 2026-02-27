import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'Industrial Equipment Marketplace',
  description: 'A platform for industrial equipment trading',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body>{children}</body>
    </html>
  )
}
