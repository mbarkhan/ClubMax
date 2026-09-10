"use client";

import Link from "next/link";
import {use, useEffect, useState} from "react";
import {useRouter} from "next/navigation";
import "../../../catalog.css";

const API = "";
type Order = {id:string; gym_id:string; status:"created"|"pending"|"succeeded"|"failed"; total_amount:number; ref_number?:string};

export default function PaymentResult({params}:{params:Promise<{locale:string}>}) {
  const {locale} = use(params), router = useRouter(), [id,setID] = useState<string|null>(null), [order,setOrder] = useState<Order|null>(null), [error,setError] = useState("");
  useEffect(()=>setID(new URLSearchParams(window.location.search).get("order")||""),[]);
  useEffect(()=>{if(id===null)return;const token=localStorage.getItem("clubmax_access_token")||"";if(!id){setError("پرداخت تکمیل نشد یا توسط کاربر لغو شد.");return}(async()=>{const headers={Authorization:`Bearer ${token}`};let response=await fetch(`${API}/v1/payments/orders/${id}`,{headers}),data=await response.json();if(response.status===401||response.status===403){router.replace(`/${locale}/login`);return}if(!response.ok)throw new Error(data.message||"دریافت نتیجه پرداخت انجام نشد.");if(data.status!=="succeeded"){const verify=await fetch(`${API}/v1/payments/orders/${id}/verify`,{method:"POST",headers});if(verify.ok)data=await verify.json();else{response=await fetch(`${API}/v1/payments/orders/${id}`,{headers});if(response.ok)data=await response.json()}}setOrder(data)})().catch(reason=>setError(reason instanceof Error?reason.message:"دریافت نتیجه پرداخت انجام نشد."))},[id,locale,router]);
  if((id===null)||(!order&&!error))return <main className="catalog-page"><div className="screen-loader"><span>در حال بررسی نتیجه پرداخت...</span></div></main>;
  const succeeded=order?.status==="succeeded",pending=order?.status==="pending"||order?.status==="created";
  return <main className="catalog-page"><section className="payment-success"><b>{succeeded?"✓":pending?"…":"!"}</b><h1>{succeeded?"پرداخت با موفقیت ثبت شد":pending?"پرداخت در حال بررسی است":"پرداخت ناموفق بود"}</h1>{succeeded?<><p>مبلغ <strong>{Number(order.total_amount).toLocaleString("fa-IR")} ریال</strong> با موفقیت تأیید شد.</p>{order.ref_number&&<p>شماره پیگیری: <strong dir="ltr">{order.ref_number}</strong></p>}</>:<p>{error||(pending?"نتیجه قطعی از درگاه دریافت نشده است؛ چند دقیقه دیگر دوباره این صفحه را باز کنید.":"پرداخت توسط درگاه تأیید نشد یا لغو شد.")}</p>}<Link href={order?`/${locale}/explore-gyms/${order.gym_id}`:`/${locale}/explore-gyms`}>{succeeded?"مشاهده باشگاه":"بازگشت به باشگاه‌ها"}</Link></section></main>;
}
