# =======================
# INTRO
# =======================
# Most of this is taken from Claude-Code's devcontainer setup
# See https://github.com/anthropics/claude-code/tree/main/.devcontainer
# We've customised this somewhat to be more suitable to our environment

FROM buildpack-deps:bookworm

ENV CLAUDE_CONFIG_DIR=/claude
ARG CLAUDE_CODE_VERSION=latest

# Install dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
  less \
  git \
  sudo \
  curl \
  fzf \
  zsh \
  unzip \
  zip \
  gnupg2 \
  gh \
  vim \
  python3 \
  python3-pip \
  python3-venv \
  wget \
  && apt-get clean && rm -rf /var/lib/apt/lists/*

# Add our user with restricted privileges
ARG USERNAME=developer
RUN groupadd --gid 1000 $USERNAME \
  && useradd --uid 1000 --gid $USERNAME --shell /bin/bash --create-home $USERNAME

# Persist bash history.
RUN SNIPPET="export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  && mkdir /commandhistory \
  && touch /commandhistory/.bash_history \
  && chown -R $USERNAME /commandhistory

# Create the Claude config directory
RUN mkdir $CLAUDE_CONFIG_DIR && chown $USERNAME $CLAUDE_CONFIG_DIR

# Install git-delta to help Claude display diffs
ARG GIT_DELTA_VERSION=0.18.2
RUN ARCH=$(dpkg --print-architecture) && \
  wget "https://github.com/dandavison/delta/releases/download/${GIT_DELTA_VERSION}/git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" && \
  sudo dpkg -i "git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb" && \
  rm "git-delta_${GIT_DELTA_VERSION}_${ARCH}.deb"

# Switch to user
USER $USERNAME

ENV SHELL=/bin/zsh

# Set the default editor and visual
ARG DEFAULT_EDITOR=vim
ENV EDITOR=$DEFAULT_EDITOR
ENV VISUAL=$DEFAULT_EDITOR

# setup ZSH and set theme etc
ARG ZSH_IN_DOCKER_VERSION=1.2.0
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v${ZSH_IN_DOCKER_VERSION}/zsh-in-docker.sh)" -- \
  -p git \
  -p fzf \
  -a "source /usr/share/doc/fzf/examples/key-bindings.zsh" \
  -a "source /usr/share/doc/fzf/examples/completion.zsh" \
  -a "export PROMPT_COMMAND='history -a' && export HISTFILE=/commandhistory/.bash_history" \
  -x

# Install python UV dependency
RUN curl -LsSf https://astral.sh/uv/install.sh | sh
ENV PATH="/home/$USERNAME/.local/bin:${PATH}"

# Install Claude
RUN curl -fsSL https://claude.ai/install.sh | bash

# Install NVM
ARG NVM_VERSION=0.40.4
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh | PROFILE=/home/$USERNAME/.zshrc bash

# Install node & yarn
ARG NODE_VERSION=lts
RUN bash -c "source /home/$USERNAME/.nvm/nvm.sh && nvm install $NODE_VERSION && npm install --global yarn"

# Install pnpm
RUN wget -qO- https://get.pnpm.io/install.sh | ENV="$HOME/.zshrc" SHELL="$(which sh)" sh -

# {{PROFILE_EXTENSIONS}}

ENTRYPOINT ["/bin/zsh"]
