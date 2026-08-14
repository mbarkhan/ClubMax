import {redirect} from "next/navigation";
export default async function LegacyMemberLogin({params}:{params:Promise<{locale:string}>}) { const {locale}=await params; redirect(`/${locale}/login`); }
