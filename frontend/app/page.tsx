export default function Home() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-between p-24">
      <div className="z-10 w-full max-w-5xl items-center justify-between font-mono text-sm lg:flex">
        <h1 className="text-4xl font-bold">Industrial Equipment Marketplace</h1>
      </div>
      <div className="relative flex place-items-center before:absolute before:h-[300px] before:w-[480px] before:-translate-x-1/2 before:rounded-full before:bg-gradient-radial before:from-white before:to-transparent before:blur-2xl before:content-[''] after:absolute after:-z-20 after:h-[180px] after:w-[240px] after:translate-x-1/3 after:bg-gradient-conic after:from-blue-200 after:via-blue-50 after:blur-2xl after:content-[''] before:bg-gradient-to-br before:from-transparent before:to-blue-700 before:opacity-10 after:from-sky-900 after:via-blue-900 after:opacity-40 before:lg:h-[360px] sm:block">
        <p className="text-2xl">Welcome to the platform</p>
      </div>
      <div className="mb-32 grid text-center lg:max-w-5xl lg:w-full lg:mb-0 lg:grid-cols-4 lg:text-left">
        <a
          href="https://nextjs.org/docs"
          className="group rounded-lg border border-transparent px-5 py-4 transition-colors hover:border-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800/30"
        >
          <h2 className="mb-3 text-2xl font-semibold">Documentation</h2>
          <p className="m-0 max-w-[30ch] text-sm opacity-50 group-hover:opacity-100">
            Learn about Next.js features and API.
          </p>
        </a>

        <a
          href="https://vercel.com/templates?framework=next.js"
          className="group rounded-lg border border-transparent px-5 py-4 transition-colors hover:border-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800/30"
        >
          <h2 className="mb-3 text-2xl font-semibold">Templates</h2>
          <p className="m-0 max-w-[30ch] text-sm opacity-50 group-hover:opacity-100">
            Discover and deploy boilerplate example Next.js projects.
          </p>
        </a>

        <a
          href="https://vercel.com/new?utm_source=create-next-app&utm_medium=appdir-template&utm_campaign=create-next-app"
          className="group rounded-lg border border-transparent px-5 py-4 transition-colors hover:border-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800/30"
        >
          <h2 className="mb-3 text-2xl font-semibold">Deploy</h2>
          <p className="m-0 max-w-[30ch] text-sm opacity-50 group-hover:opacity-100">
            Instantly deploy your Next.js site to a shareable URL with Vercel.
          </p>
        </a>
      </div>
    </main>
  )
}
