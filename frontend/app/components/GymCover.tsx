"use client";

import {useEffect, useState} from "react";

const API = "";

export default function GymCover({gymId}: {gymId: string}) {
  const [url,setURL] = useState("");
  useEffect(()=>{
    fetch(`${API}/v1/public/gyms/${gymId}/images`)
      .then(response=>response.ok?response.json():[])
      .then(images=>setURL(images?.[0]?.storage_key?`${API}/media/${images[0].storage_key}`:""))
      .catch(()=>setURL(""));
  },[gymId]);
  return <div className="gym-cover" aria-label={url?"تصویر باشگاه":"باشگاه بدون تصویر"} style={url?{backgroundImage:`url(${url})`}:undefined}/>;
}
