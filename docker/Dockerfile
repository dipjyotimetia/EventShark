FROM docker.vectorized.io/vectorized/redpanda:latest

# Copy scripts
COPY create_topics.sh /create_topics.sh
COPY schema/expense.json /expense.json
COPY schema/payment.json /payment.json
COPY schema/transaction.json /transaction.json

WORKDIR /

USER root
RUN chown root:root /create_topics.sh
# Make scripts executable
RUN chmod +x /create_topics.sh

# Set the entrypoint to create_topics.sh
ENTRYPOINT ["/create_topics.sh"]