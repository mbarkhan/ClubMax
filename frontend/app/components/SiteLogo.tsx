import Image from "next/image";
import Link from "next/link";

type SiteLogoProps = {
  href: string;
  className?: string;
  priority?: boolean;
};

export default function SiteLogo({href, className = "logo", priority = false}: SiteLogoProps) {
  return (
    <Link href={href} className={className} aria-label="GyMax - صفحه اصلی">
      <Image
        src="/images/gymax-logo.png"
        alt="GyMax"
        width={1983}
        height={793}
        priority={priority}
        sizes="(max-width: 768px) 112px, 132px"
      />
    </Link>
  );
}
