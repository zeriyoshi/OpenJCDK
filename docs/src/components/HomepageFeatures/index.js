import clsx from 'clsx';
import Heading from '@theme/Heading';
import styles from './styles.module.css';

const FeatureList = [
  {
    title: '低コスト',
    description: (
      <>
        邪神ちゃん画像botは Google Cloud のマネージド サービスを活用して構築されており、月あたり約 10 円程度と超低コストで運用されています。
      </>
    ),
  },
  {
    title: '公式許諾済み',
    description: (
      <>
        邪神ちゃん画像botは公式アカウントから投稿されており、版権 (著作権) に関する法的な問題がクリアです。
      </>
    ),
  },
  {
    title: 'オープンソース',
    description: (
      <>
        邪神ちゃん画像botのソースおよびドキュメントは GitHub ですべて公開されており、誰でも閲覧・貢献することができます。
      </>
    ),
  },
];

function Feature({Svg, title, description}) {
  return (
    <div className={clsx('col col--4')}>
      <div className="text--center padding-horiz--md">
        <Heading as="h3">{title}</Heading>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
