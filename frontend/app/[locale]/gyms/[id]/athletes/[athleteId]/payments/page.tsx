"use client";

import Link from "next/link";
import {use,useEffect,useMemo,useState} from "react";
import {useRouter} from "next/navigation";
import "../../../../../../payments.css";

const API=process.env.NEXT_PUBLIC_API_URL!;
type Athlete={id:string;first_name:string;last_name:string};
type Payment={id:string;gym_name:string;tuition_amount:number;platform_fee:number;total_amount:number;status:string;tracking_id:string;billing_year?:number;billing_month?:number;payment_kind:"registration"|"monthly";created_at:string};
const money=(value:number)=>`${Number(value).toLocaleString("fa-IR")} ریال`;

export default function AthletePayments({params}:{params:Promise<{locale:string;id:string;athleteId:string}>}){
  const{locale,id,athleteId}=use(params),router=useRouter(),[athlete,setAthlete]=useState<Athlete|null>(null),[payments,setPayments]=useState<Payment[]>([]),[error,setError]=useState("");
  useEffect(()=>{const token=localStorage.getItem("gymax_access_token")||"",headers={Authorization:`Bearer ${token}`,"Accept-Language":"fa"};Promise.all([fetch(`${API}/v1/gyms/${id}/athletes`,{headers}),fetch(`${API}/v1/gyms/${id}/athletes/${athleteId}/payments`,{headers})]).then(async([members,transactions])=>{if(members.status===401||members.status===403||transactions.status===401||transactions.status===403){router.replace(`/${locale}/login`);return}const memberData=await members.json(),paymentData=await transactions.json();if(!members.ok)throw new Error(memberData.message);if(!transactions.ok)throw new Error(paymentData.message);setAthlete((memberData||[]).find((item:Athlete)=>item.id===athleteId)||null);setPayments(paymentData||[])}).catch(reason=>setError(reason instanceof Error?reason.message:"دریافت پرداخت‌ها انجام نشد."))},[athleteId,id,locale,router]);
  const total=useMemo(()=>payments.reduce((sum,item)=>sum+Number(item.tuition_amount),0),[payments]);
  return <main className="catalog-page coach-athlete-payments"><header><Link href={`/${locale}/gyms/${id}/athletes`}>بازگشت به فهرست ورزشکاران ←</Link></header><section className="wallet-head"><span>ریز شهریه‌های واریزی</span><h1>{athlete?`${athlete.first_name} ${athlete.last_name}`:"پرداخت‌های ورزشکار"}</h1><strong>{money(total)}</strong></section>{error&&<div className="alert">{error}</div>}<section className="payments-list"><div className="payments-heading"><div><span>سوابق مالی این باشگاه</span><h2>فهرست شهریه‌های پرداخت‌شده</h2></div><small>{payments.length.toLocaleString("fa-IR")} تراکنش</small></div>{payments.map(payment=><article key={payment.id}><div className="payment-main"><span className="payment-status">پرداخت موفق</span><h3>{payment.gym_name}</h3></div><dl><div><dt>نوع پرداخت</dt><dd>{payment.payment_kind==="monthly"?"شهریه ماهیانه":"ثبت‌نام در باشگاه"}</dd></div><div><dt>دوره شهریه</dt><dd>{payment.billing_year&&payment.billing_month?`${payment.billing_year.toLocaleString("fa-IR",{useGrouping:false})}/${payment.billing_month.toLocaleString("fa-IR")}`:"—"}</dd></div><div><dt>مبلغ شهریه</dt><dd>{money(payment.tuition_amount)}</dd></div><div><dt>کارمزد پلتفرم</dt><dd>{money(payment.platform_fee)}</dd></div><div><dt>تاریخ پرداخت</dt><dd>{new Date(payment.created_at).toLocaleString("fa-IR")}</dd></div><div className="tracking-cell"><dt>کد پیگیری</dt><dd dir="ltr">{payment.tracking_id}</dd></div></dl></article>)}{!payments.length&&!error&&<div className="payments-empty"><b>پرداختی برای این ورزشکار ثبت نشده است.</b></div>}</section></main>;
}
