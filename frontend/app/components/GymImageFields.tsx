"use client";

import FileField from "./FileField";

type ExistingImage = {id: string; url: string};

export default function GymImageFields({existing = [], onDelete}: {existing?: ExistingImage[]; onDelete?: (id:string)=>Promise<void>}) {
  const available = Math.max(0, 3 - existing.length);
  return <section className="gym-image-manager">
    {existing.length > 0 && <div className="existing-gym-images"><span>تصاویر فعلی</span><div>{existing.map(image=><figure key={image.id}><img src={image.url} alt="تصویر باشگاه"/>{onDelete&&<button type="button" onClick={()=>{if(confirm("این تصویر حذف شود؟"))void onDelete(image.id)}} aria-label="حذف تصویر" title="حذف تصویر">×</button>}</figure>)}</div></div>}
    {available > 0 ? <div className="single-image-fields">
      {Array.from({length:available},(_,index)=><FileField key={index} name="images" label={`تصویر ${existing.length+index+1}`} maxFiles={1} maxSizeKB={500} accept="image/jpeg,image/png,image/webp" className="gym-images-field"/>)}
    </div> : <p className="image-limit-note">هر سه تصویر باشگاه ثبت شده‌اند.</p>}
  </section>;
}
