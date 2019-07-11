import * as React from "react";
import PropTypes from "prop-types";
import { compose, withApollo } from "react-apollo";
import { withRouter, Link } from "react-router-dom";
import ShipLoading from "./ShipLoading";
import { getWatchById, getParentWatch } from "../queries/WatchQueries";
import { Utilities } from "../utilities/utilities";
import Loader from "./shared/Loader";

import "../scss/components/ShipCompleted.scss"

export class ShipInitCompleted extends React.Component {
  static propTypes = {
    initSessionId: PropTypes.string,
    onActiveInitSessionCompleted: PropTypes.func.isRequired,
  }

  state = {
    loadingWatch: true,
    watchSlug: "",
    watchId: "",
    isLoading: false
  };

  async componentDidMount() {
    this.interval = setInterval(async () => await this.queryForWatchByID(), 1000);
  }

  componentWillUnmount() {
    clearInterval(this.interval);
  }

  queryForWatchByID = async () => {
    const { client, initSessionId } = this.props;

    if (!initSessionId) {
      this.props.history.push("/watches");
    }

    const { data } = await client.query({
      query: getWatchById,
      variables: { id: initSessionId },
      // TODO errors are thrown for nonexistent watch
      fetchPolicy: "network-only",
      errorPolicy: "ignore",
    });

    if (data.getWatch && data.getWatch.watchName) {
      clearInterval(this.interval);
      await this.props.refetchListWatches();
      if (data.getWatch.cluster) {
        const parentResult = await client.query({
          query: getParentWatch,
          variables: { id: initSessionId },
          fetchPolicy: "network-only"
        });
        return this.props.history.push(`/watch/${parentResult.data.getParentWatch.slug}/downstreams`);
      }
      this.setState({ loadingWatch: false, watchSlug: data.getWatch.slug, watchId: initSessionId });
    }

    return;
  }

  handleDownload = async () => {
    const { initSessionId } = this.props;
    this.setState({ isLoading: true });
    await Utilities.handleDownload(initSessionId);
    this.setState({ isLoading: false });
  }

  handleGotoWatches = () => {
    const { history, onActiveInitSessionCompleted } = this.props;
    const { watchSlug } = this.state;
    onActiveInitSessionCompleted();
    history.push(`/watch/${watchSlug}/downstreams?add=1`);
  }

  render() {
    const { loadingWatch, isLoading } = this.state;

    if (loadingWatch) {
      return (
        <div className="Login-wrapper container flex-column flex1 u-overflow--auto">
          <ShipLoading headerText="Finalizing application" subText="Give us a second to cross the t's and dot the i's." />
        </div>
      )
    }

    return (
      <div className="Login-wrapper container flex-column flex1 u-overflow--auto">
        <div className="flex-column flex1 alignItems--center justifyContent--center">
          <div className="init-complete-wrapper flex-auto">
            <div className="flex1 flex-column u-textAlign--center">
              <p className="u-fontSize--larger u-color--tuna u-fontWeight--bold u-lineHeight--normal">Your application is ready to be deployed, what's next?</p>
              <p className="u-fontSize--normal u-color--dustyGray u-fontWeight--medium u-lineHeight--normal u-marginTop--5 u-marginBottom--30">Now that your yaml is ready to be deployed, you can configure PRs to be made automatically to GitHub so that you can easily deploy it. You can also download the rendered assets to deploy and test on your servers.</p>
            </div>

            <div className="u-flexTabletReflow u-paddingBottom--20 ship-complete-integration-cards-wrapper justifyContent--center flexWrap--wrap">
              <div className="ship-complete-integration-card-wrapper flex-auto u-paddingBottom--20">
                <div className="flex-column flex1 ship-complete-integration-card">
                  <div className="flex-column flex1 justifyContent--center alignItems--center">
                    <span className="icon ship-complete-yml-dl u-marginTop--10"></span>
                    <p className="u-color--tundora u-fontWeight--bold u-textAlign--center u-fontSize--normal u-marginTop--20">Download YAML</p>
                  </div>
                  <div className="u-marginTop--10">
                    <p className="u-fontSize--small u-fontWeight--medium u-textAlign--center u-lineHeight--normal u-color--dustyGray">You can download the YAML generated by Replicated Ship to deploy &amp; test on your server.</p>
                  </div>
                  <div className="button-wrapper flex">
                    {isLoading ?
                      <div className="flex-column flex1 alignItems--center justifyContent--center">
                        <Loader size="60" />
                      </div>
                      :
                      <div className="flex1 flex card-action-wrapper u-cursor--pointer u-textAlign--center">
                        <span className="flex1 card-action u-color--astral u-fontSize--small u-fontWeight--medium" onClick={this.handleDownload}>Download deployment YAML</span>
                      </div>
                    }
                  </div>
                </div>
              </div>
              <div className="ship-complete-integration-card-wrapper flex-auto u-paddingBottom--20">
                <div className="flex-column flex1 ship-complete-integration-card">
                  <div className="flex alignItems--center justifyContent--center u-marginTop--10">
                    <span className="icon ship-complete-icon-gh"></span>
                    <span className="deployment-or-text">OR</span>
                    <span className="icon ship-medium-size"></span>
                  </div>
                  <div className="u-textAlign--center">
                    <p className="u-color--tundora u-fontWeight--bold u-textAlign--center u-fontSize--normal u-marginTop--20">Deploy to a cluster</p>
                    <p className="u-fontSize--small u-fontWeight--medium u-textAlign--center u-lineHeight--normal u-color--dustyGray u-marginTop--10">Select one of your existing clusters to get started with deployments.</p>
                  </div>
                  <div className="button-wrapper flex">
                    <div className="flex1 flex card-action-wrapper u-cursor--pointer u-textAlign--center">
                      <span className="flex1 card-action u-color--astral u-fontSize--small u-fontWeight--medium" onClick={this.handleGotoWatches}>Add to a deployment cluster</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div className="u-textAlign--center">
              <p className="u-fontSize--small u-color--dustyGray u-lineHeight--normal u-fontWeight--medium">Not sure what you want to do? You can <Link to="/watches" className="replicated-link">head back to your watches dashboard</Link> and decide later.</p>
            </div>
          </div>
        </div>
      </div>
    );
  }
}

export default compose(
  withRouter,
  withApollo,
)(ShipInitCompleted);
