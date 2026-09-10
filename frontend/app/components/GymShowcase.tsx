"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

const API = "";
const SLIDE_DURATION = 4000;

type ShowcaseGym = {
  id: string;
  name: string;
  coach_name: string;
  sport: string;
  sport_style: string;
  start_time: string;
  end_time: string;
  image_storage_key?: string;
};

const shortTime = (value: string) => value?.slice(0, 5) || "—";

export default function GymShowcase({ locale }: { locale: string }) {
  const [gyms, setGyms] = useState<ShowcaseGym[]>([]);
  const [active, setActive] = useState(0);
  const [loading, setLoading] = useState(true);
  const [paused, setPaused] = useState(false);

  useEffect(() => {
    const controller = new AbortController();
    fetch(`${API}/v1/gyms/showcase`, { signal: controller.signal, cache: "no-store" })
      .then(async (response) => {
        if (!response.ok) throw new Error("showcase request failed");
        return response.json();
      })
      .then((items: ShowcaseGym[]) => setGyms(Array.isArray(items) ? items : []))
      .catch((error: unknown) => {
        if (!(error instanceof DOMException && error.name === "AbortError")) setGyms([]);
      })
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, []);

  useEffect(() => {
    if (gyms.length < 2 || paused) return;
    const timer = window.setTimeout(
      () => setActive((current) => (current + 1) % gyms.length),
      SLIDE_DURATION
    );
    return () => window.clearTimeout(timer);
  }, [active, gyms.length, paused]);

  const show = (index: number) => {
    setActive((index + gyms.length) % gyms.length);
  };

  if (loading) {
    return (
      <section id="gyms" className="gym-showcase gym-showcase-loading" aria-busy="true">
        <div className="showcase-skeleton" />
      </section>
    );
  }

  if (!gyms.length) {
    return (
      <section id="gyms" className="gym-showcase gym-showcase-empty">
        <div>
          <span>باشگاه‌های ClubMax</span>
          <h2>جای باشگاه شما اینجاست</h2>
          <p>باشگاه خود را ثبت کنید تا در ویترین ClubMax دیده شوید.</p>
          <Link href={`/${locale}/signup`}>ثبت باشگاه در ClubMax</Link>
        </div>
      </section>
    );
  }

  const gym = gyms[active];
  const imageURL = gym.image_storage_key ? `${API}/media/${gym.image_storage_key}` : "";

  return (
    <section
      id="gyms"
      className={`gym-showcase${paused ? " paused" : ""}`}
      aria-roledescription="اسلایدر"
      aria-label="باشگاه‌های ثبت‌شده در ClubMax"
      onMouseEnter={() => setPaused(true)}
      onMouseLeave={() => setPaused(false)}
      onFocusCapture={() => setPaused(true)}
      onBlurCapture={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) setPaused(false);
      }}
    >
      <article className="showcase-slide" key={gym.id}>
        <div className="showcase-media">
          {imageURL ? (
            <img
              src={imageURL}
              alt={`نمای باشگاه ${gym.name}`}
              onError={(event) => {
                event.currentTarget.hidden = true;
              }}
            />
          ) : null}
          <span className="showcase-shade" />
        </div>
        <div className="showcase-content">
          <span className="showcase-kicker">باشگاه‌های ثبت‌شده در ClubMax</span>
          <h2>{gym.name}</h2>
          <p className="showcase-sport">
            {gym.sport} <b>·</b> {gym.sport_style}
          </p>
          <dl>
            <div>
              <dt>مربی</dt>
              <dd>{gym.coach_name}</dd>
            </div>
            <div>
              <dt>ساعت فعالیت</dt>
              <dd dir="ltr">
                {shortTime(gym.start_time)} – {shortTime(gym.end_time)}
              </dd>
            </div>
          </dl>
          <Link href={`/${locale}/explore-gyms`}>مشاهده باشگاه‌ها</Link>
        </div>
      </article>

      {gyms.length > 1 ? (
        <>
          <div className="showcase-arrows">
            <button type="button" onClick={() => show(active - 1)} aria-label="باشگاه قبلی">
              ‹
            </button>
            <button type="button" onClick={() => show(active + 1)} aria-label="باشگاه بعدی">
              ›
            </button>
          </div>
          <div className="showcase-dots" aria-label="انتخاب باشگاه">
            <span>{`${active + 1} / ${gyms.length}`}</span>
            {gyms.length <= 12
              ? gyms.map((item, index) => (
                  <button
                    type="button"
                    key={item.id}
                    className={index === active ? "active" : ""}
                    onClick={() => show(index)}
                    aria-label={`نمایش ${item.name}`}
                    aria-current={index === active ? "true" : undefined}
                  />
                ))
              : null}
          </div>
          <span key={active} className="showcase-progress" />
        </>
      ) : null}
    </section>
  );
}
