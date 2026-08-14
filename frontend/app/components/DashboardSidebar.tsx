"use client";

import Link from "next/link";
import {useRouter} from "next/navigation";
import SiteLogo from "./SiteLogo";
import "../dashboard-sidebar.css";

const API=process.env.NEXT_PUBLIC_API_URL!;
type SidebarUser={role:"admin"|"coach"|"athlete";first_name:string;last_name:string;face_image_id?:string};
const roleNames={admin:"مدیر سایت",coach:"مربی",athlete:"ورزشکار"};

export default function DashboardSidebar({user,locale,active}:{user:SidebarUser;locale:string;active:"dashboard"|"payments"|"profile"}){
  const router=useRouter(),initials=user.first_name[0]+(user.last_name?.[0]||""),profileImage=user.face_image_id?`${API}/media/${user.face_image_id}`:"";
  const item=(key:typeof active,href:string,label:string)=><Link className={active===key?"active":""} href={href}>{label}</Link>;
  return <aside><SiteLogo href={`/${locale}`} priority/><div className="mobile-dashboard-links"><Link className={active==="dashboard"?"active":""} href={`/${locale}/dashboard`}>داشبورد</Link>{user.role==="athlete"&&<Link className={active==="payments"?"active":""} href={`/${locale}/payments`}>پرداخت‌ها</Link>}<Link className={active==="profile"?"active":""} href={`/${locale}/profile`}>پروفایل</Link></div><div className="identity"><div className="profile-thumbnail">{profileImage?<img src={profileImage} alt={`عکس ${user.first_name} ${user.last_name}`}/>:initials}</div><b>{user.first_name} {user.last_name}</b><span>{roleNames[user.role]}</span></div><nav>{item("dashboard",`/${locale}/dashboard`,"داشبورد")}{user.role==="admin"&&<a>کاربران</a>}{user.role==="coach"&&<Link href={`/${locale}/gyms`}>باشگاه‌های من</Link>}{user.role==="coach"&&<Link href={`/${locale}/wallet`}>کیف پول و تراکنش‌ها</Link>}{user.role==="athlete"&&<Link href={`/${locale}/explore-gyms`}>باشگاه‌ها</Link>}{user.role==="athlete"&&item("payments",`/${locale}/payments`,"پرداخت‌های من")}<a>پیام‌ها</a>{item("profile",`/${locale}/profile`,"پروفایل من")}</nav><button className="logout" onClick={()=>{localStorage.clear();router.replace(`/${locale}/login`)}}>خروج از حساب</button></aside>;
}
