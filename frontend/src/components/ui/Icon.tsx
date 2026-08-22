import type { SVGProps } from 'react'

/**
 * Inline icon set.
 *
 * Bundled as components rather than an icon font or sprite so stroke weight and
 * sizing stay consistent and nothing extra is downloaded.
 */
type IconProps = SVGProps<SVGSVGElement>

function Svg({ children, ...props }: IconProps) {
  return (
    <svg
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      focusable="false"
      width="1em"
      height="1em"
      {...props}
    >
      {children}
    </svg>
  )
}

export const ArrowUpRight = (props: IconProps) => (
  <Svg {...props}>
    <path d="M7 17 17 7M9 7h8v8" />
  </Svg>
)

export const ArrowDownRight = (props: IconProps) => (
  <Svg {...props}>
    <path d="M7 7l10 10M17 9v8H9" />
  </Svg>
)

export const ArrowRight = (props: IconProps) => (
  <Svg {...props}>
    <path d="M5 12h14M13 6l6 6-6 6" />
  </Svg>
)

export const ArrowLeft = (props: IconProps) => (
  <Svg {...props}>
    <path d="M19 12H5M11 18l-6-6 6-6" />
  </Svg>
)

export const ChevronDown = (props: IconProps) => (
  <Svg {...props}>
    <path d="m6 9 6 6 6-6" />
  </Svg>
)

export const ChevronRight = (props: IconProps) => (
  <Svg {...props}>
    <path d="m9 6 6 6-6 6" />
  </Svg>
)

export const Close = (props: IconProps) => (
  <Svg {...props}>
    <path d="M18 6 6 18M6 6l12 12" />
  </Svg>
)

export const Check = (props: IconProps) => (
  <Svg {...props}>
    <path d="m20 6-11 11-5-5" />
  </Svg>
)

export const Sparkle = (props: IconProps) => (
  <Svg {...props}>
    <path d="M12 3v4M12 17v4M5.5 6.5l2.5 2.5M16 15l2.5 2.5M3 12h4M17 12h4M5.5 17.5 8 15M16 9l2.5-2.5" />
    <circle cx="12" cy="12" r="2.4" />
  </Svg>
)

export const Warning = (props: IconProps) => (
  <Svg {...props}>
    <path d="M12 4.5 3 19.5h18L12 4.5Z" />
    <path d="M12 10v4.5M12 17.2v.3" />
  </Svg>
)

export const Info = (props: IconProps) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="8.5" />
    <path d="M12 11v5.5M12 7.8v.3" />
  </Svg>
)

export const Clock = (props: IconProps) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="8.5" />
    <path d="M12 7.5V12l3 2" />
  </Svg>
)

export const Refresh = (props: IconProps) => (
  <Svg {...props}>
    <path d="M20 11a8 8 0 1 0-2.6 6.4" />
    <path d="M20 5v6h-6" />
  </Svg>
)

export const ExternalLink = (props: IconProps) => (
  <Svg {...props}>
    <path d="M14 5h5v5M19 5l-7 7" />
    <path d="M18 14v4a1.5 1.5 0 0 1-1.5 1.5h-10A1.5 1.5 0 0 1 5 18V7.5A1.5 1.5 0 0 1 6.5 6H10" />
  </Svg>
)

export const Send = (props: IconProps) => (
  <Svg {...props}>
    <path d="M4.5 12 20 5l-6.5 15-2.2-6.3L4.5 12Z" />
  </Svg>
)

export const Stop = (props: IconProps) => (
  <Svg {...props}>
    <rect x="7" y="7" width="10" height="10" rx="2" />
  </Svg>
)

export const Wallet = (props: IconProps) => (
  <Svg {...props}>
    <rect x="3.5" y="6" width="17" height="13" rx="2.5" />
    <path d="M3.5 10h17M16 14.5h1.5" />
  </Svg>
)

export const Chart = (props: IconProps) => (
  <Svg {...props}>
    <path d="M4 19V5M4 19h16" />
    <path d="m7 15 3.5-4 3 2.5L20 7" />
  </Svg>
)

export const Layers = (props: IconProps) => (
  <Svg {...props}>
    <path d="m12 4 8 4-8 4-8-4 8-4Z" />
    <path d="m4 12 8 4 8-4M4 16l8 4 8-4" />
  </Svg>
)

export const Scenario = (props: IconProps) => (
  <Svg {...props}>
    <path d="M5 19V9M5 19h14" />
    <path d="M9 15V7M13 19v-6M17 15V5" />
  </Svg>
)

export const Copy = (props: IconProps) => (
  <Svg {...props}>
    <rect x="9" y="9" width="11" height="11" rx="2" />
    <path d="M15 6.5V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v7a2 2 0 0 0 2 2h.5" />
  </Svg>
)

export const Shield = (props: IconProps) => (
  <Svg {...props}>
    <path d="M12 3.5 5 6v6c0 4 3 7 7 8.5 4-1.5 7-4.5 7-8.5V6l-7-2.5Z" />
    <path d="m9.5 12 1.8 1.8L15 10" />
  </Svg>
)

export const Menu = (props: IconProps) => (
  <Svg {...props}>
    <path d="M4 7h16M4 12h16M4 17h16" />
  </Svg>
)

export const User = (props: IconProps) => (
  <Svg {...props}>
    <circle cx="12" cy="9" r="3.5" />
    <path d="M5 20c1.2-3.2 3.8-4.8 7-4.8s5.8 1.6 7 4.8" />
  </Svg>
)

export const Question = (props: IconProps) => (
  <Svg {...props}>
    <circle cx="12" cy="12" r="8.5" />
    <path d="M9.8 9.5a2.2 2.2 0 1 1 3.4 1.9c-.7.5-1.2.9-1.2 1.9M12 16.4v.3" />
  </Svg>
)
