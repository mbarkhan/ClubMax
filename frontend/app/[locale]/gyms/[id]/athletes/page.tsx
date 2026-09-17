"use client";

import Link from "next/link";
import {use,useEffect,useState} from "react";
import {useRouter} from "next/navigation";
import "../../../../catalog.css";
import "../../../../roster.css";

const API="";
type Athlete={id:string;first_name:string;last_name:string;phone?:string;age:number;tuition_status:"settled"|"debtor";unpaid_months:number;outstanding_amount:number;registered_at:string};

export default function GymAthletes({params}:{params:Promise<{locale:string;id:string}>}){
  const{locale,id}=use(params),router=useRouter(),[items,setItems]=useState<Athlete[]>([]),[error,setError]=useState("");
  useEffect(()=>{const token=localStorage.getItem("clubmax_access_token")||"";fetch(`${API}/v1/gyms/${id}/athletes`,{headers:{Authorization:`Bearer ${token}`}}).then(async response=>{if(response.status===401||response.status===403){router.replace(`/${locale}/login`);return}const data=await response.json();if(!response.ok)throw new Error(data.message);setItems(data||[])}).catch(reason=>setError(reason instanceof Error?reason.message:"دریافت ورزشکاران انجام نشد."))},[id,locale,router]);
  return <main className="catalog-page"><header><Link href={`/${locale}/gyms`}>بازگشت به باشگاه‌های من ←</Link></header><section className="wallet-ledger roster-card"><div className="roster-head"><div><span>اعضای باشگاه</span><h1>ورزشکاران ثبت‌نام‌شده</h1></div><strong>{items.length.toLocaleString("fa-IR")} ورزشکار</strong></div>{error&&<div className="alert">{error}</div>}<div className="roster-table"><div className="roster-row roster-title"><span>نام و نام خانوادگی</span><span>شماره موبایل</span><span>سن</span><span>وضعیت شهریه</span><span>تاریخ ثبت‌نام</span></div>{items.map(item=><div className="roster-row" key={item.id}><b>{item.first_name} {item.last_name}</b><span dir="ltr">{item.phone||"—"}</span><span>{item.age?`${item.age.toLocaleString("fa-IR")} سال`:"—"}</span><Link className={`tuition-cell ${item.tuition_status}`} href={`/${locale}/gyms/${id}/athletes/${item.id}/payments`}>{item.tuition_status==="debtor"?<><b>بدهکار</b><small>{item.outstanding_amount.toLocaleString("fa-IR")} ریال · {item.unpaid_months.toLocaleString("fa-IR")} ماه</small></>:<><b>تسویه‌شده</b><small>مشاهده پرداخت‌ها</small></>}</Link><time>{new Date(item.registered_at).toLocaleDateString("fa-IR")}</time></div>)}{!items.length&&!error&&<div className="roster-empty">هنوز ورزشکاری در این باشگاه ثبت‌نام نکرده است.</div>}</div></section></main>;
}
