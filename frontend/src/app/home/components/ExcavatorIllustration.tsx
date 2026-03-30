interface ExcavatorIllustrationProps {
  seed?: number;
}

const SKIES: [string, string][] = [
  ["#87CEEB", "#b8d8ea"],
  ["#7ab8d4", "#a8cce0"],
  ["#9ec8de", "#c0d9e8"],
];

export function ExcavatorIllustration({ seed = 0 }: ExcavatorIllustrationProps) {
  const [skyTop, skyBot] = SKIES[seed % SKIES.length];
  const armRotate = [-25, -20, -22][seed % 3];
  const bucketX   = [270, 275, 272][seed % 3];
  const id        = `sky-${seed}`;

  return (
    <svg
      viewBox="0 0 400 200"
      xmlns="http://www.w3.org/2000/svg"
      className="w-full h-full block"
      aria-hidden="true"
    >
      <defs>
        <linearGradient id={id} x1="0%" y1="0%" x2="0%" y2="100%">
          <stop offset="0%"   stopColor={skyTop} />
          <stop offset="100%" stopColor={skyBot} />
        </linearGradient>
      </defs>

      {/* Sky */}
      <rect width="400" height="200" fill={`url(#${id})`} />

      {/* Clouds */}
      <ellipse cx={60  + seed * 20} cy="38" rx="44" ry="20" fill="rgba(255,255,255,0.88)" />
      <ellipse cx={95  + seed * 20} cy="32" rx="32" ry="22" fill="rgba(255,255,255,0.88)" />
      <ellipse cx={310 - seed * 10} cy="46" rx="48" ry="22" fill="rgba(255,255,255,0.85)" />
      <ellipse cx={345 - seed * 10} cy="41" rx="30" ry="19" fill="rgba(255,255,255,0.85)" />

      {/* Ground */}
      <rect x="0" y="155" width="400" height="45" fill="#c9a86c" />
      <rect x="0" y="155" width="400" height="8"  fill="#b8965a" />

      {/* Body */}
      <rect x="120" y="110" width="145" height="53" rx="8" fill="#F5A623" />
      <rect x="125" y="105" width="93"  height="30" rx="6" fill="#d89010" />
      {/* Cab glass */}
      <rect x="130" y="110" width="37"  height="22" rx="3" fill="#c8e8f5" />
      {/* Stripe */}
      <rect x="120" y="135" width="145" height="6"  fill="#d89010" />

      {/* Boom arm */}
      <rect
        x="240" y="72" width="15" height="72" rx="5"
        fill="#c88010"
        transform={`rotate(${armRotate} 240 72)`}
      />
      {/* Stick arm */}
      <rect
        x="264" y="58" width="13" height="66" rx="5"
        fill="#F5A623"
        transform={`rotate(${-armRotate / 2} 264 58)`}
      />
      {/* Bucket */}
      <path
        d={`M${bucketX} 116 Q${bucketX + 22} 110 ${bucketX + 26} 132 Q${bucketX + 16} 142 ${bucketX} 137 Z`}
        fill="#2a2a2a"
      />

      {/* Tracks */}
      <rect x="110" y="158" width="165" height="18" rx="9" fill="#333" />
      <rect x="115" y="162" width="155" height="10" rx="5" fill="#555" />
      {[130, 157, 188, 218, 248].map((cx, i) => (
        <circle key={i} cx={cx} cy="167" r={i === 0 || i === 4 ? 10 : 8} fill="#222" />
      ))}
    </svg>
  );
}
