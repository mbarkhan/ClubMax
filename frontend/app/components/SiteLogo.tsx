import Image from "next/image";
import Link from "next/link";

type SiteLogoProps = {
  href: string;
  className?: string;
  priority?: boolean;
};

export default function SiteLogo({href, className = "logo", priority = false}: SiteLogoProps) {
  return (
    <Link href={href} className={className} aria-label="ClubMax - صفحه اصلی">
      <Image
        src="/images/clubmax-logo.svg"
        alt="ClubMax"
        width={500}
        height={180}
        priority={priority}
        sizes="(max-width: 768px) 112px, 132px"
      />
    </Link>
  );
}
