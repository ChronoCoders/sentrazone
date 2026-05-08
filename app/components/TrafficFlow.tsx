const nodes = [
  {
    id: "device",
    label: "Your Device",
    sublabel: "Origin",
    cx: 80,
    icon: (
      <g transform="translate(80,80)">
        <rect x="-11" y="-8" width="22" height="15" rx="2" fill="none" stroke="#C9A84C" strokeWidth="1.5" />
        <rect x="-14" y="7" width="28" height="2.5" rx="1" fill="#C9A84C" opacity="0.6" />
        <circle cx="0" cy="-1" r="2" fill="#C9A84C" opacity="0.5" />
      </g>
    ),
  },
  {
    id: "tunnel",
    label: "Encrypted Tunnel",
    sublabel: "End-to-End",
    cx: 260,
    icon: (
      <g transform="translate(260,80)">
        <rect x="-7" y="-2" width="14" height="10" rx="2" fill="none" stroke="#C9A84C" strokeWidth="1.5" />
        <path d="M-5,-2 v-4 a5,5 0 0,1 10,0 v4" fill="none" stroke="#C9A84C" strokeWidth="1.5" strokeLinecap="round" />
        <circle cx="0" cy="3" r="1.5" fill="#C9A84C" opacity="0.7" />
      </g>
    ),
  },
  {
    id: "infra",
    label: "Private Infrastructure",
    sublabel: "Your Servers",
    cx: 450,
    icon: (
      <g transform="translate(450,80)">
        {[{y:-8},{y:0},{y:8}].map((row, i) => (
          <g key={i}>
            <rect x="-12" y={row.y - 3} width="24" height="6" rx="1" fill="none" stroke="#C9A84C" strokeWidth="1.2" />
            <circle cx="8" cy={row.y} r="1.2" fill="#C9A84C" opacity="0.7" />
          </g>
        ))}
      </g>
    ),
  },
  {
    id: "exit",
    label: "Residential Exit",
    sublabel: "City-Specific",
    cx: 640,
    icon: (
      <g transform="translate(640,80)">
        <path d="M0,-12 L12,0 L12,12 L-12,12 L-12,0 Z" fill="none" stroke="#C9A84C" strokeWidth="1.5" strokeLinejoin="round" />
        <rect x="-4" y="4" width="8" height="8" rx="1" fill="none" stroke="#C9A84C" strokeWidth="1.2" />
      </g>
    ),
  },
  {
    id: "clean",
    label: "Clean IP",
    sublabel: "Target Destination",
    cx: 820,
    icon: (
      <g transform="translate(820,80)">
        <circle cx="0" cy="0" r="12" fill="none" stroke="#C9A84C" strokeWidth="1.5" />
        <path d="M-6,0 L-2,4 L7,-5" fill="none" stroke="#C9A84C" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      </g>
    ),
  },
];

const lineSegments = [
  { x1: 112, x2: 228 },
  { x1: 292, x2: 418 },
  { x1: 482, x2: 608 },
  { x1: 672, x2: 788 },
];

export default function TrafficFlow() {
  return (
    <div className="relative w-full overflow-x-auto">
      <svg
        viewBox="0 0 900 190"
        className="w-full min-w-[600px]"
        aria-label="Traffic flow: Your Device to Encrypted Tunnel to Private Infrastructure to Residential Exit to Clean IP"
      >
        <defs>
          <filter id="glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="4" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <filter id="node-glow" x="-50%" y="-50%" width="200%" height="200%">
            <feGaussianBlur in="SourceGraphic" stdDeviation="8" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
          <linearGradient id="line-grad" x1="0%" y1="0%" x2="100%" y2="0%">
            <stop offset="0%" stopColor="#C9A84C" stopOpacity="0.1" />
            <stop offset="50%" stopColor="#C9A84C" stopOpacity="0.5" />
            <stop offset="100%" stopColor="#C9A84C" stopOpacity="0.1" />
          </linearGradient>
        </defs>

        {/* Connecting lines */}
        {lineSegments.map((seg, i) => (
          <g key={i}>
            <line
              x1={seg.x1} y1="80"
              x2={seg.x2} y2="80"
              stroke="#242424"
              strokeWidth="1.5"
            />
            <line
              x1={seg.x1} y1="80"
              x2={seg.x2} y2="80"
              stroke="#C9A84C"
              strokeWidth="1"
              strokeDasharray="4 8"
              opacity="0.35"
            >
              <animate
                attributeName="stroke-dashoffset"
                from="12"
                to="0"
                dur="1.5s"
                repeatCount="indefinite"
              />
            </line>
          </g>
        ))}

        {/* Animated packet */}
        <circle r="5" fill="#C9A84C" filter="url(#glow)" opacity="0.9">
          <animateMotion
            dur="3.5s"
            repeatCount="indefinite"
            calcMode="linear"
            path="M80,80 L260,80 L450,80 L640,80 L820,80"
          />
          <animate
            attributeName="opacity"
            values="0;1;1;1;0"
            keyTimes="0;0.05;0.92;0.97;1"
            dur="3.5s"
            repeatCount="indefinite"
          />
        </circle>

        {/* Packet trail glow */}
        <circle r="10" fill="#C9A84C" filter="url(#glow)" opacity="0.2">
          <animateMotion
            dur="3.5s"
            repeatCount="indefinite"
            calcMode="linear"
            path="M80,80 L260,80 L450,80 L640,80 L820,80"
          />
          <animate
            attributeName="opacity"
            values="0;0.2;0.2;0.2;0"
            keyTimes="0;0.05;0.92;0.97;1"
            dur="3.5s"
            repeatCount="indefinite"
          />
        </circle>

        {/* Node circles */}
        {nodes.map((node) => (
          <g key={node.id}>
            {/* Outer glow ring */}
            <circle cx={node.cx} cy="80" r="34" fill="#C9A84C" opacity="0.04" />
            {/* Main circle */}
            <circle
              cx={node.cx} cy="80" r="28"
              fill="#111111"
              stroke="#2a2a2a"
              strokeWidth="1.5"
            />
            {/* Gold accent ring */}
            <circle
              cx={node.cx} cy="80" r="28"
              fill="none"
              stroke="#C9A84C"
              strokeWidth="1"
              opacity="0.3"
            />
            {/* Icon */}
            {node.icon}
          </g>
        ))}

        {/* Labels */}
        {nodes.map((node) => (
          <g key={node.id + "-label"}>
            <text
              x={node.cx}
              y="128"
              textAnchor="middle"
              fill="#ededed"
              fontSize="11"
              fontFamily="var(--font-geist-sans)"
              fontWeight="500"
            >
              {node.label}
            </text>
            <text
              x={node.cx}
              y="144"
              textAnchor="middle"
              fill="#888888"
              fontSize="9.5"
              fontFamily="var(--font-geist-sans)"
            >
              {node.sublabel}
            </text>
          </g>
        ))}
      </svg>
    </div>
  );
}
