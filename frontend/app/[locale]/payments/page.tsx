"use client";

import {use,useEffect,useMemo,useState} from "react";
import {useRouter} from "next/navigation";
import DashboardSidebar from "../../components/DashboardSidebar";
import "../../payments.css";

const API="";
type User={role:"athlete";first_name:string;last_name:string;face_image_id?:string};
type Payment={id:string;gym_id:string;gym_name:string;coach_name:string;tuition_amount:number;platform_fee:number;total_amount:number;status:string;tracking_id:string;billing_year?:number;billing_month?:number;payment_kind:"registration"|"monthly";created_at:string};
const money=(value:number)=>`${Number(value).toLocaleString("fa-IR")} ریال`;

export default function Payments({params}:{params:Promise<{locale:string}>}){
  const{locale}=use(params),router=useRouter(),[user,setUser]=useState<User|null>(null),[payments,setPayments]=useState<Payment[]>([]),[error,setError]=useState("");
  useEffect(()=>{const token=localStorage.getItem("clubmax_access_token")||"",headers={Authorization:`Bearer ${token}`,"Accept-Language":"fa"};Promise.all([fetch(`${API}/v1/users/me`,{headers}),fetch(`${API}/v1/payments/me`,{headers})]).then(async([profile,list])=>{if(profile.status===401||profile.status===403||list.status===401||list.status===403){router.replace(`/${locale}/login`);return}const profileData=await profile.json(),paymentData=await list.json();if(!profile.ok)throw new Error(profileData.message);if(profileData.role!=="athlete"){router.replace(`/${locale}/dashboard`);return}if(!list.ok)throw new Error(paymentData.message);setUser(profileData);setPayments(paymentData||[])}).catch(reason=>setError(reason instanceof Error?reason.message:"دریافت پرداخت‌ها انجام نشد."))},[locale,router]);
  const total=useMemo(()=>payments.reduce((sum,item)=>sum+Number(item.total_amount),0),[payments]);
  if(!user)return <div className="screen-loader"><span>{error||"در حال دریافت پرداخت‌ها..."}</span></div>;
  return <div className="shell athlete-shell"><DashboardSidebar user={user} locale={locale} active="payments"/><main className="dash payments-page"><header><div><span>سوابق مالی</span><h1>پرداخت‌های من</h1></div></header><section className="payment-summary"><article><span>تعداد پرداخت‌ها</span><strong>{payments.length.toLocaleString("fa-IR")}</strong></article><article><span>مجموع پرداخت‌شده</span><strong>{money(total)}</strong></article></section>{error&&<div className="alert">{error}</div>}<section className="payments-list"><div className="payments-heading"><div><span>تراکنش‌های موفق</span><h2>فهرست کامل پرداخت‌ها</h2></div><small>جدیدترین پرداخت در ابتدای فهرست قرار دارد.</small></div>{payments.map(payment=><article key={payment.id}><div className="payment-main"><span className="payment-status">پرداخت موفق</span><h3>{payment.gym_name}</h3><p>مربی: {payment.coach_name}</p></div><dl><div><dt>نوع پرداخت</dt><dd>{payment.payment_kind==="monthly"?"شهریه ماهیانه":"ثبت‌نام در باشگاه"}</dd></div><div><dt>دوره شهریه</dt><dd>{payment.billing_year&&payment.billing_month?`${payment.billing_year.toLocaleString("fa-IR",{useGrouping:false})}/${payment.billing_month.toLocaleString("fa-IR")}`:"—"}</dd></div><div><dt>مبلغ شهریه</dt><dd>{money(payment.tuition_amount)}</dd></div><div><dt>کارمزد</dt><dd>{money(payment.platform_fee)}</dd></div><div><dt>مبلغ کل</dt><dd className="payment-total">{money(payment.total_amount)}</dd></div><div><dt>تاریخ پرداخت</dt><dd>{new Date(payment.created_at).toLocaleString("fa-IR")}</dd></div><div className="tracking-cell"><dt>کد پیگیری</dt><dd dir="ltr">{payment.tracking_id}</dd></div></dl></article>)}{!payments.length&&!error&&<div className="payments-empty"><b>هنوز پرداختی ثبت نشده است</b><p>پس از ثبت‌نام یا پرداخت شهریه، تراکنش‌ها در این صفحه نمایش داده می‌شوند.</p></div>}</section></main></div>;
}
