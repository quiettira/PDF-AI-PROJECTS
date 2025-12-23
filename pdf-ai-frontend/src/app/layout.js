export default function RootLayout({ children }) {
  return (
    <html lang="en">
      <body style={{ backgroundColor: "#ffffff", color: "#000000" }}>
        {children}
      </body>
    </html>
  );
}
