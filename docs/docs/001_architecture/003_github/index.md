# GitHub

[GitHub Actions](https://docs.github.com/en/actions) は GitHub 上で利用できる CD/CI サービスです。 OpenJCDK では実行するアプリケーション コンテナ イメージを GitHub Actions を用いてビルドしています。

GitHub Actions によってビルドされたコンテナ イメージは Google Cloud の [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation) を用いて Google Cloud 上に push されます。

アプリケーション コンテナ ビルドを行うGitHub Actions の構成は `.github/workflows/deploy_app.yaml` に記述されています。

GitHub Actions には以下のリポジトリ シークレットが設定されています。

- `GCP_SERVICE_ACCOUNT` Google Cloud のサービス アカウント キー
- `GCP_WORKLOAD_IDENTITY_PROVIDER` Google Cloud Workflow Identity のプロバイダ

## GitHub Actions の構成

:::warning
このドキュメントはまだ完成していません
:::

### GitHub Actions Secret の構成 

:::warning
このドキュメントはまだ完成していません
:::
